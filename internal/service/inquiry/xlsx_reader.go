package inquiry

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

type xlsxSheet struct {
	Name string
	Rows [][]string
}

type xlsxInlineText struct {
	Value string `xml:",chardata"`
}

type xlsxInlineStr struct {
	T []xlsxInlineText `xml:"t"`
}

type xlsxCell struct {
	Ref    string        `xml:"r,attr"`
	Type   string        `xml:"t,attr"`
	Value  string        `xml:"v"`
	Inline xlsxInlineStr `xml:"is"`
}

type xlsxRow struct {
	Cells []xlsxCell `xml:"c"`
}

type xlsxWorksheet struct {
	Rows []xlsxRow `xml:"sheetData>row"`
}

func readXLSXSheets(data []byte) ([]xlsxSheet, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("无法读取压缩文件: %w", err)
	}
	files := map[string]*zip.File{}
	for _, f := range reader.File {
		files[f.Name] = f
	}
	workbookFile, ok := files["xl/workbook.xml"]
	if !ok {
		return nil, fmt.Errorf("缺少 workbook.xml")
	}
	workbookData, err := readZipFile(workbookFile)
	if err != nil {
		return nil, err
	}
	workbook, err := parseWorkbook(workbookData)
	if err != nil {
		return nil, err
	}
	relsFile, ok := files["xl/_rels/workbook.xml.rels"]
	if !ok {
		return nil, fmt.Errorf("缺少 workbook 关联文件")
	}
	relsData, err := readZipFile(relsFile)
	if err != nil {
		return nil, err
	}
	rels, err := parseRelationships(relsData)
	if err != nil {
		return nil, err
	}
	var sharedStrings []string
	if ssFile, ok := files["xl/sharedStrings.xml"]; ok {
		ssData, err := readZipFile(ssFile)
		if err != nil {
			return nil, err
		}
		sharedStrings, err = parseSharedStrings(ssData)
		if err != nil {
			return nil, err
		}
	}
	var sheets []xlsxSheet
	for _, s := range workbook {
		target, ok := rels[s.RID]
		if !ok {
			return nil, fmt.Errorf("未找到 sheet 对应关系: %s", s.Name)
		}
		path := path.Clean("xl/" + target)
		sheetFile, ok := files[path]
		if !ok {
			return nil, fmt.Errorf("缺少 sheet 文件: %s", path)
		}
		sheetData, err := readZipFile(sheetFile)
		if err != nil {
			return nil, err
		}
		rows, err := parseWorksheet(sheetData, sharedStrings)
		if err != nil {
			return nil, fmt.Errorf("解析 sheet 失败 (%s): %w", s.Name, err)
		}
		sheets = append(sheets, xlsxSheet{Name: s.Name, Rows: rows})
	}
	return sheets, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

type workbookSheet struct {
	Name string
	RID  string
}

func parseWorkbook(data []byte) ([]workbookSheet, error) {
	type sheetXML struct {
		Name string `xml:"name,attr"`
		RID  string `xml:"r:id,attr"`
	}
	type workbookXML struct {
		Sheets []sheetXML `xml:"sheets>sheet"`
	}
	var wb workbookXML
	if err := xml.Unmarshal(data, &wb); err != nil {
		return nil, fmt.Errorf("解析 workbook 失败: %w", err)
	}
	var sheets []workbookSheet
	for _, s := range wb.Sheets {
		if strings.TrimSpace(s.Name) == "" || strings.TrimSpace(s.RID) == "" {
			continue
		}
		sheets = append(sheets, workbookSheet{Name: s.Name, RID: s.RID})
	}
	return sheets, nil
}

func parseRelationships(data []byte) (map[string]string, error) {
	type relXML struct {
		ID     string `xml:"Id,attr"`
		Target string `xml:"Target,attr"`
	}
	type relsXML struct {
		Relationships []relXML `xml:"Relationship"`
	}
	var rels relsXML
	if err := xml.Unmarshal(data, &rels); err != nil {
		return nil, fmt.Errorf("解析关系文件失败: %w", err)
	}
	out := make(map[string]string, len(rels.Relationships))
	for _, r := range rels.Relationships {
		out[r.ID] = r.Target
	}
	return out, nil
}

func parseSharedStrings(data []byte) ([]string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var result []string
	var builder strings.Builder
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("解析 sharedStrings 失败: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "si" {
				builder.Reset()
			} else if t.Name.Local == "t" {
				var text string
				if err := dec.DecodeElement(&text, &t); err != nil {
					return nil, fmt.Errorf("解析 sharedStrings 文本失败: %w", err)
				}
				builder.WriteString(text)
			}
		case xml.EndElement:
			if t.Name.Local == "si" {
				result = append(result, builder.String())
			}
		}
	}
	return result, nil
}

func parseWorksheet(data []byte, sharedStrings []string) ([][]string, error) {
	var sheet xlsxWorksheet
	if err := xml.Unmarshal(data, &sheet); err != nil {
		return nil, fmt.Errorf("解析 worksheet 失败: %w", err)
	}
	var rows [][]string
	for _, row := range sheet.Rows {
		var maxIdx int
		for _, cell := range row.Cells {
			idx := columnIndex(cell.Ref)
			if idx > maxIdx {
				maxIdx = idx
			}
		}
		line := make([]string, maxIdx+1)
		currentIdx := 0
		for _, cell := range row.Cells {
			idx := columnIndex(cell.Ref)
			if idx < 0 {
				idx = currentIdx
			}
			currentIdx = idx + 1
			val := resolveCellValue(cell, sharedStrings)
			if idx >= len(line) {
				tmp := make([]string, idx+1)
				copy(tmp, line)
				line = tmp
			}
			line[idx] = val
		}
		rows = append(rows, line)
	}
	return rows, nil
}

func resolveCellValue(cell xlsxCell, sharedStrings []string) string {
	if len(cell.Inline.T) > 0 {
		var builder strings.Builder
		for _, t := range cell.Inline.T {
			builder.WriteString(t.Value)
		}
		return builder.String()
	}
	switch cell.Type {
	case "s":
		idx, err := strconv.Atoi(strings.TrimSpace(cell.Value))
		if err == nil && idx >= 0 && idx < len(sharedStrings) {
			return sharedStrings[idx]
		}
	case "str", "inlineStr":
		return cell.Value
	}
	return cell.Value
}

func columnIndex(ref string) int {
	ref = strings.ToUpper(ref)
	i := 0
	for i < len(ref) && ref[i] >= 'A' && ref[i] <= 'Z' {
		i++
	}
	if i == 0 {
		return -1
	}
	letters := ref[:i]
	idx := 0
	for j := 0; j < len(letters); j++ {
		idx = idx*26 + int(letters[j]-'A'+1)
	}
	return idx - 1
}
