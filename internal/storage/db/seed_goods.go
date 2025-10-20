package foodDB

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	category "hdzk.cn/foodapp/internal/domain/category"
	dict "hdzk.cn/foodapp/internal/domain/dict"
	"hdzk.cn/foodapp/internal/domain/goods"
	"hdzk.cn/foodapp/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 商品基础数据（根据默认品类、规格预制）
var defaultGoods = []struct {
	Name               string
	CategoryName       string
	SpecName           string
	AcceptanceStandard string
}{
	// —— 肉类 —— //
	{"猪前腿肉", "肉类", "冷鲜", "色泽淡红、肌理紧实、表面微干不黏手、无异味、按压能回弹、无淤血与淋巴组织"},
	{"猪后腿肉", "肉类", "冷鲜", "色泽均匀、脂肪洁白、肌肉有弹性、无注水、无异味"},
	{"五花肉", "肉类", "冷鲜", "肥瘦相间、纹理清晰、无酸败异味、无发黏"},
	{"排骨", "肉类", "冷鲜", "骨面洁净、肉质紧实、无异味、无血水外渗"},
	{"猪肝", "肉类", "冷鲜", "色泽红褐、切面湿润不黏、无异味、无结节斑点"},
	{"猪心", "肉类", "冷鲜", "外表完整、色泽正常、无异味、无大量血水"},
	{"猪肚", "肉类", "冷鲜", "清洗干净、无异味、无破损、组织有弹性"},
	{"牛腱子", "肉类", "冷鲜", "颜色鲜红、筋膜清晰、组织紧致、无异味"},
	{"肥牛片", "肉类", "冷冻", "切片整齐、肥瘦分明、无酸败异味、冻结均匀无解冻回冻"},
	{"牛上脑", "肉类", "冷鲜", "大理石纹理明显、肉香正常、无异味、回弹良好"},
	{"羊肉卷", "肉类", "冷冻", "色泽红润、油花均匀、无酸败、无解冻回冻痕迹"},
	{"鸡整只", "肉类", "冷鲜", "皮肤完整、色泽正常、无毛根、无异味、胸肌饱满"},
	{"鸡腿", "肉类", "冷鲜", "肉质紧实、皮肤无破损、无异味、无血水渗出"},
	{"鸡胸肉", "肉类", "冷鲜", "颜色淡粉、纤维细致、无异味、表面微干不黏"},
	{"鸭肉", "肉类", "冷鲜", "皮面干爽、气味正常、无黏液、组织有弹性"},

	// —— 干货 —— //
	{"木耳（干）", "干货", "散装", "朵形完整、色黑发亮、无杂质泥沙、无霉变异味、无虫蛀"},
	{"香菇（干）", "干货", "散装", "菇盖完整、柄短肉厚、气味纯正、无霉点、无硫味"},
	{"海带结（干）", "干货", "散装", "色泽棕褐、无异味、无砂土杂质、无霉变"},
	{"紫菜（干）", "干货", "袋装", "片张完整、干燥、无霉点、无杂质、无异味"},
	{"粉丝", "干货", "袋装", "色泽透明或乳白、无霉点、无异味、无结块"},
	{"粉条（土豆/红薯）", "干货", "袋装", "条形完整、干燥无结块、无异味、无霉变"},
	{"虾皮（干）", "干货", "散装", "色泽自然、体表清洁、无霉味、无砂石杂质"},
	{"枸杞（干）", "干货", "袋装", "粒形完整、色泽均匀、无硫味、无虫蛀、干燥适中"},
	{"花生米（生）", "干货", "散装", "颗粒饱满、无霉变、无虫蛀、无异味"},
	{"红枣（干）", "干货", "袋装", "个大皮皱、果肉致密、无破损霉变、无异味"},

	// —— 牛奶 —— //
	{"纯牛奶", "牛奶", "盒装", "包装完好、在保质期内、无鼓包泄漏、外观洁净、标签标识齐全"},
	{"酸奶", "牛奶", "盒装", "冷藏运输、口味正常、包装完好无膨胀、在保质期内"},
	{"学生奶（低温）", "牛奶", "盒装", "2-6℃冷链、包装完好、标签清晰、无鼓包"},
	{"炼乳/淡奶油", "牛奶", "罐装", "罐体完好无胀罐锈蚀、在保质期内、无异味"},

	// —— 豆类淀粉制品 —— //
	{"黄豆（干）", "豆类淀粉制品", "散装", "颗粒饱满、色泽均匀、无霉变虫蛀、无异味"},
	{"绿豆（干）", "豆类淀粉制品", "散装", "粒形完整、无霉点、无杂质、无异味"},
	{"红豆（干）", "豆类淀粉制品", "散装", "颗粒均匀饱满、无霉变虫蛀"},
	{"黑豆（干）", "豆类淀粉制品", "散装", "表皮光洁、无霉变、无杂质异味"},
	{"玉米淀粉", "豆类淀粉制品", "袋装", "干燥细腻、无结块、无异味、包装完好"},
	{"红薯淀粉", "豆类淀粉制品", "袋装", "粉质洁白、无结块结潮、无异味"},
	{"土豆淀粉", "豆类淀粉制品", "袋装", "粉质细腻、干燥、无杂质、包装完好"},
	{"藕粉", "豆类淀粉制品", "袋装", "粉末均匀、干燥、无霉味、标签齐全"},

	// —— 调料 —— //
	{"食用盐", "调料", "袋装", "包装完好、标签齐全、干燥无结块、无杂质"},
	{"白砂糖", "调料", "袋装", "颗粒均匀、干燥无结块、无异味"},
	{"酱油", "调料", "瓶装", "色泽正常、无浑浊沉淀异常、包装完好、在保质期内"},
	{"食醋", "调料", "瓶装", "酸香纯正、无杂质浑浊、包装完好在保质期内"},
	{"料酒", "调料", "瓶装", "气味纯正、无浑浊沉淀异常、包装完好"},
	{"蚝油", "调料", "瓶装", "浓稠度正常、无胀瓶渗漏、标签齐全、在保质期内"},
	{"豆瓣酱", "调料", "瓶装/袋装", "色泽红亮、气味纯正、无霉变、包装完好"},
	{"黄豆酱", "调料", "袋装", "气味纯正、无霉点、包装完好、在保质期内"},
	{"芝麻酱", "调料", "瓶装", "香味纯正、无酸败、无霉味、包装完好"},
	{"花椒", "调料", "袋装", "干燥、香气浓郁、无霉变虫蛀、无杂质"},
	{"干辣椒/辣椒面", "调料", "袋装", "色泽鲜红、干燥无霉点、气味纯正无异味"},
	{"胡椒粉", "调料", "袋装", "细度均匀、香味纯正、无霉变结块"},
	{"八角/桂皮", "调料", "袋装", "干燥度好、香味纯正、无霉变虫蛀"},
	{"鸡精/味精", "调料", "袋装", "颗粒均匀、干燥无结块、包装完好"},

	// —— 果蔬 —— //
	{"青椒", "果蔬", "新鲜", "果面光洁、色泽青绿、肉质硬实、无虫蛀、无腐烂、无异味"},
	{"红椒", "果蔬", "新鲜", "表皮鲜亮、无皱缩软烂、无虫斑、无异味"},
	{"西红柿", "果蔬", "新鲜", "果形端正、色泽红润、无裂纹软烂、无异味"},
	{"黄瓜", "果蔬", "新鲜", "表面有刺新鲜、瓜身直挺、无软烂、无异味"},
	{"茄子", "果蔬", "新鲜", "皮色乌亮、果柄青绿、无皱缩、无虫咬、无腐烂"},
	{"土豆", "果蔬", "新鲜", "块大匀称、表皮光滑、无发芽青皮、无霉烂"},
	{"胡萝卜", "果蔬", "新鲜", "根条完整、色橙红、质地脆嫩、无裂口腐烂"},
	{"白萝卜", "果蔬", "新鲜", "根体完整、肉质紧实、无空心、无腐烂异味"},
	{"洋葱", "果蔬", "新鲜", "外皮干燥完整、球体坚实、无发芽霉变"},
	{"大葱", "果蔬", "新鲜", "叶鞘洁净、含水正常、无枯萎、无异味"},
	{"生姜", "果蔬", "新鲜", "块大致密、纤维少、无干瘪霉变、无异味"},
	{"大蒜", "果蔬", "新鲜", "蒜头饱满、瓣粒紧实、无发芽、无霉变"},
	{"蒜薹", "果蔬", "新鲜", "色泽青绿、脆嫩挺直、无老化、无腐烂"},
	{"西兰花", "果蔬", "新鲜", "花球紧密、色泽浓绿、无黄花、无腐烂异味"},
	{"菜花", "果蔬", "新鲜", "花球紧密洁白、无散花、无腐烂异味"},
	{"娃娃菜", "果蔬", "新鲜", "叶球紧实、色泽嫩黄、无抽苔、无腐烂"},
	{"大白菜", "果蔬", "新鲜", "叶片完整、帮脆白、无抽苔烂心、无异味"},
	{"圆白菜", "果蔬", "新鲜", "球形紧实、叶片无虫洞霉斑、无腐烂"},
	{"生菜", "果蔬", "新鲜", "叶片翠绿、无泥沙杂质、无腐烂抽苔"},
	{"油麦菜", "果蔬", "新鲜", "叶片笔直、色泽青绿、无老化、无腐烂"},
	{"菠菜", "果蔬", "新鲜", "叶片完整、根部新鲜、无黄叶、无泥沙"},
	{"香菜", "果蔬", "新鲜", "香味浓郁、叶片完整、无枯黄烂叶"},
	{"芹菜", "果蔬", "新鲜", "茎秆脆嫩、叶色青绿、无倒伏、无腐烂"},
	{"空心菜", "果蔬", "新鲜", "茎叶鲜嫩、色泽正常、无泥沙、无腐烂"},
	{"韭菜", "果蔬", "新鲜", "叶片宽厚、气味清香、无抽苔黄叶、无腐烂"},
	{"四季豆/豆角", "果蔬", "新鲜", "豆荚饱满、无鼓籽老化、无斑疤、无腐烂"},
	{"豇豆", "果蔬", "新鲜", "条形均匀、质地脆嫩、无虫蛀、无腐烂"},
	{"冬瓜", "果蔬", "新鲜", "表面蜡粉自然、瓜身坚实、无碰伤腐烂"},
	{"南瓜", "果蔬", "新鲜", "外皮完整、成熟度适中、无破损霉斑"},
	{"西葫芦", "果蔬", "新鲜", "表皮光洁、手感坚挺、无软烂、无异味"},
	{"苦瓜", "果蔬", "新鲜", "瓜体挺直、瓜瘤整齐、无黄化软烂"},
	{"藕", "果蔬", "新鲜", "节断面洁白、孔洞干净、无泥腥异味、无腐烂"},
	{"玉米（鲜）", "果蔬", "新鲜", "苞叶青绿、须丝金黄、颗粒饱满、无虫蛀霉烂"},
	{"香菇（鲜）", "果蔬", "新鲜", "菇盖完整、菌柄洁白、气味清香、无黏滑霉变"},
	{"平菇", "果蔬", "新鲜", "菌盖饱满、组织致密、无异味、无霉变"},
	{"金针菇", "果蔬", "新鲜", "菌柄洁白细长、无腐烂异味、包装清洁"},
	{"木耳（鲜发）", "果蔬", "新鲜", "色泽黑亮、耳片完整、无异味、无杂质"},
	{"苹果", "果蔬", "新鲜", "果形端正、表皮光洁、无碰伤腐烂、无蜡味"},
	{"香蕉", "果蔬", "新鲜", "果皮金黄、果柄青绿、无压伤、无异味"},
	{"橙子", "果蔬", "新鲜", "果皮光洁、手感沉实、无霉斑、无腐烂"},
	{"梨", "果蔬", "新鲜", "果面洁净、硬度适中、无黑斑腐烂"},
	{"葡萄", "果蔬", "新鲜", "果粒饱满、果粉自然、无脱粒霉烂"},
	{"西瓜", "果蔬", "新鲜", "瓜纹清晰、敲声清脆、无裂纹腐烂"},
	{"哈密瓜", "果蔬", "新鲜", "网纹清晰、香味自然、无破损腐烂"},

	// —— 鲜（冻）水产品 —— //
	{"鲤鱼", "鲜（冻）水产品", "鲜活/冰鲜", "眼球清亮、鳃丝鲜红、体表粘液透明、腹部紧实无异味"},
	{"草鱼", "鲜（冻）水产品", "鲜活/冰鲜", "鳃红、眼亮、鱼体完整、无异味"},
	{"鲫鱼", "鲜（冻）水产品", "鲜活/冰鲜", "鱼鳞完整、眼球凸亮、鳃色红润、无异味"},
	{"鲈鱼", "鲜（冻）水产品", "冰鲜", "鱼体挺直、腹部紧实、无异味、无破损"},
	{"带鱼", "鲜（冻）水产品", "冷冻", "体表银亮、肉质紧实、无解冻回冻、无异味"},
	{"黄花鱼", "鲜（冻）水产品", "冷冻", "体形完整、鱼鳞紧密、无异味、无回冻痕"},
	{"鲳鱼", "鲜（冻）水产品", "冷冻", "体表洁净、肉质紧实、无异味、无解冻回冻"},
	{"鱿鱼圈/须", "鲜（冻）水产品", "冷冻", "肉质洁白、无氨味、无回冻、包装完整"},
	{"对虾", "鲜（冻）水产品", "冷冻", "体壳完整、弯曲自然、无黑头、无氨味"},
	{"河虾", "鲜（冻）水产品", "冰鲜", "体色透明发亮、活力好或冰鲜无异味、无死亡腐败"},
	{"虾仁", "鲜（冻）水产品", "冷冻", "色泽洁白、无黑变、无杂质、无回冻"},
	{"鲢鱼/青鱼切块", "鲜（冻）水产品", "冷冻", "切面新鲜、无血水过多、无异味"},

	// —— 蛋类 —— //
	{"鸡蛋", "蛋类", "散装", "壳面清洁完整、无破裂、无异味、摇动无晃荡声"},
	{"鸭蛋", "蛋类", "散装", "蛋壳完整、无裂纹、无异味"},
	{"鹌鹑蛋", "蛋类", "盒装", "壳面干净、大小均匀、无破损、在保质期内"},
	{"皮蛋", "蛋类", "袋装/散装", "外包膜完整、无异味、切面凝固均匀、在保质期内"},
	{"咸鸭蛋", "蛋类", "袋装/散装", "壳面干净、无渗漏异味、在保质期内"},

	// —— 豆制品 —— //
	{"老豆腐（北豆腐）", "豆制品", "散装", "块形完整、组织紧实、气味清香豆香、无酸败粘滑"},
	{"内酯豆腐（嫩豆腐）", "豆制品", "盒装", "包装完好、凝固均匀、无鼓包、在保质期内"},
	{"豆腐干", "豆制品", "散装/袋装", "块形规整、干香、无霉点、无酸败异味"},
	{"千叶豆腐", "豆制品", "袋装", "组织细腻、弹性好、无异味、包装完好"},
	{"豆皮/腐竹", "豆制品", "袋装", "色泽金黄、干燥无霉点、复水后无异味"},
	{"豆花", "豆制品", "散装", "组织细腻、豆香纯正、无酸败、无杂质"},

	// —— 面筋制品 —— //
	{"油面筋", "面筋制品", "散装/袋装", "球体饱满、组织蓬松、无哈喇味、无异味"},
	{"烤麸", "面筋制品", "散装/袋装", "组织孔洞均匀、无酸败异味、无霉点"},
	{"素鸡/素肉", "面筋制品", "袋装", "成型完整、组织紧实、气味正常、包装完好"},
}

func toPtr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func collectUnique(values []string) []string {
	uniq := make(map[string]struct{}, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		uniq[v] = struct{}{}
	}
	out := make([]string, 0, len(uniq))
	for k := range uniq {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func loadSpecIDs(ctx context.Context, db *gorm.DB, names []string) (map[string]string, error) {
	if len(names) == 0 {
		return map[string]string{}, nil
	}

	var specs []dict.Spec
	if err := db.WithContext(ctx).Where("name IN ?", names).Find(&specs).Error; err != nil {
		return nil, fmt.Errorf("查询规格失败: %w", err)
	}
	result := make(map[string]string, len(specs))
	for _, sp := range specs {
		result[sp.Name] = sp.ID
	}

	for _, name := range names {
		if _, ok := result[name]; !ok {
			return nil, fmt.Errorf("未找到规格 '%s'，请先初始化 base_spec", name)
		}
	}
	return result, nil
}

func loadCategoryIDs(ctx context.Context, db *gorm.DB, names []string) (map[string]string, error) {
	if len(names) == 0 {
		return map[string]string{}, nil
	}

	var categories []category.Category
	if err := db.WithContext(ctx).Where("name IN ?", names).Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询品类失败: %w", err)
	}
	result := make(map[string]string, len(categories))
	for _, cat := range categories {
		result[cat.Name] = cat.ID
	}

	for _, name := range names {
		if _, ok := result[name]; !ok {
			return nil, fmt.Errorf("未找到品类 '%s'，请先初始化 base_category", name)
		}
	}
	return result, nil
}

func upsertGoods(ctx context.Context, db *gorm.DB, row *goods.Goods) error {
	assignments := map[string]any{
		"code":        row.Code,
		"category_id": row.CategoryID,
		"spec_id":     row.SpecID,
		"is_deleted":  0,
		"updated_at":  time.Now(),
	}
	if row.ImageURL != nil {
		assignments["image_url"] = *row.ImageURL
	} else {
		assignments["image_url"] = nil
	}
	if row.AcceptanceStandard != nil {
		assignments["acceptance_standard"] = *row.AcceptanceStandard
	} else {
		assignments["acceptance_standard"] = nil
	}

	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "org_id"}, {Name: "name"}, {Name: "spec_id"}},
			DoUpdates: clause.Assignments(assignments),
		}).
		Create(row).Error
}

func EnsureDefaultGoods(ctx context.Context, gdb *gorm.DB) error {
	specNames := make([]string, 0, len(defaultGoods))
	categoryNames := make([]string, 0, len(defaultGoods))
	for _, g := range defaultGoods {
		specNames = append(specNames, g.SpecName)
		categoryNames = append(categoryNames, g.CategoryName)
	}

	specIDs, err := loadSpecIDs(ctx, gdb, collectUnique(specNames))
	if err != nil {
		return err
	}
	categoryIDs, err := loadCategoryIDs(ctx, gdb, collectUnique(categoryNames))
	if err != nil {
		return err
	}

	success := 0
	for _, item := range defaultGoods {
		row := &goods.Goods{
			Name:               item.Name,
			OrgID:              DefaultOrgID,
			SpecID:             specIDs[item.SpecName],
			CategoryID:         categoryIDs[item.CategoryName],
			AcceptanceStandard: toPtr(item.AcceptanceStandard),
			Sort:               0,
		}
		if err := upsertGoods(ctx, gdb, row); err != nil {
			logger.L().Error("Failed to seed goods", zap.Error(err),
				zap.String("name", item.Name),
			)
			return err
		}
		success++
	}

	logger.L().Info("Goods seeded/ensured",
		zap.Int("count", success),
		zap.String("org_id", DefaultOrgID),
	)
	return nil
}
