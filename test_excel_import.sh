#!/bin/bash

# Excel导入测试脚本
# 使用方法: ./test_excel_import.sh <excel_file> <org_id> <token>

set -e

EXCEL_FILE="${1}"
ORG_ID="${2}"
TOKEN="${3}"
API_BASE="${4:-http://localhost:8080/api/v1}"

if [ -z "$EXCEL_FILE" ] || [ -z "$ORG_ID" ] || [ -z "$TOKEN" ]; then
    echo "Usage: $0 <excel_file> <org_id> <token> [api_base]"
    echo "Example: $0 market_price.xlsx 550e8400-e29b-41d4-a716-446655440000 your_token"
    exit 1
fi

if [ ! -f "$EXCEL_FILE" ]; then
    echo "Error: File $EXCEL_FILE not found"
    exit 1
fi

echo "=========================================="
echo "Excel导入测试"
echo "=========================================="
echo "文件: $EXCEL_FILE"
echo "组织ID: $ORG_ID"
echo "API地址: $API_BASE"
echo "=========================================="
echo ""

# 计算文件MD5
echo "1. 计算文件MD5..."
if command -v md5sum &> /dev/null; then
    FILE_MD5=$(md5sum "$EXCEL_FILE" | awk '{print $1}')
elif command -v md5 &> /dev/null; then
    FILE_MD5=$(md5 -q "$EXCEL_FILE")
else
    echo "Error: md5sum or md5 command not found"
    exit 1
fi
echo "   MD5: $FILE_MD5"
echo ""

# 获取文件信息
FILENAME=$(basename "$EXCEL_FILE")
FILESIZE=$(stat -f%z "$EXCEL_FILE" 2>/dev/null || stat -c%s "$EXCEL_FILE" 2>/dev/null)
CHUNK_SIZE=$((1024 * 1024)) # 1MB
TOTAL_CHUNKS=$(( ($FILESIZE + $CHUNK_SIZE - 1) / $CHUNK_SIZE ))

echo "2. 文件信息"
echo "   文件名: $FILENAME"
echo "   大小: $FILESIZE bytes"
echo "   切片数: $TOTAL_CHUNKS"
echo ""

# 上传切片
echo "3. 上传文件切片..."
for ((i=0; i<$TOTAL_CHUNKS; i++)); do
    echo "   上传切片 $i/$TOTAL_CHUNKS..."
    
    # 计算切片的起始和结束位置
    START=$((i * CHUNK_SIZE))
    
    # 提取切片
    CHUNK_FILE="/tmp/${FILENAME}.part${i}"
    dd if="$EXCEL_FILE" of="$CHUNK_FILE" bs=$CHUNK_SIZE skip=$i count=1 2>/dev/null
    
    # 上传切片
    RESPONSE=$(curl -s -X POST "${API_BASE}/excel/upload_chunk" \
        -H "Authorization: Bearer $TOKEN" \
        -F "filename=$FILENAME" \
        -F "chunk_index=$i" \
        -F "file=@$CHUNK_FILE")
    
    # 检查响应
    if echo "$RESPONSE" | grep -q '"ok":true'; then
        echo "   ✓ 切片 $i 上传成功"
    else
        echo "   ✗ 切片 $i 上传失败: $RESPONSE"
        rm -f "$CHUNK_FILE"
        exit 1
    fi
    
    # 清理临时文件
    rm -f "$CHUNK_FILE"
done
echo ""

# 合并切片
echo "4. 合并文件切片..."
MERGE_RESPONSE=$(curl -s -X POST "${API_BASE}/excel/merge_chunks" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"filename\": \"$FILENAME\",
        \"total_chunks\": $TOTAL_CHUNKS,
        \"md5\": \"$FILE_MD5\"
    }")

if echo "$MERGE_RESPONSE" | grep -q '"ok":true'; then
    echo "   ✓ 文件合并成功"
    FILEPATH=$(echo "$MERGE_RESPONSE" | grep -o '"filepath":"[^"]*"' | cut -d'"' -f4)
    echo "   文件路径: $FILEPATH"
else
    echo "   ✗ 文件合并失败: $MERGE_RESPONSE"
    exit 1
fi
echo ""

# 校验Excel
echo "5. 校验Excel文件..."
VALIDATE_RESPONSE=$(curl -s -X POST "${API_BASE}/excel/validate" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"filepath\": \"$FILEPATH\"}")

if echo "$VALIDATE_RESPONSE" | grep -q '"ok":true'; then
    echo "   ✓ Excel文件校验通过"
    echo "$VALIDATE_RESPONSE" | jq '.' 2>/dev/null || echo "$VALIDATE_RESPONSE"
else
    echo "   ✗ Excel文件校验失败: $VALIDATE_RESPONSE"
    exit 1
fi
echo ""

# 导入数据
echo "6. 导入Excel数据..."
IMPORT_RESPONSE=$(curl -s -X POST "${API_BASE}/excel/import" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"filepath\": \"$FILEPATH\",
        \"org_id\": \"$ORG_ID\"
    }")

if echo "$IMPORT_RESPONSE" | grep -q '"ok":true'; then
    echo "   ✓ Excel数据导入成功"
    echo "$IMPORT_RESPONSE" | jq '.' 2>/dev/null || echo "$IMPORT_RESPONSE"
else
    echo "   ✗ Excel数据导入失败: $IMPORT_RESPONSE"
    exit 1
fi
echo ""

echo "=========================================="
echo "测试完成！"
echo "=========================================="
