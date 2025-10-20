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
	Code               string
	CategoryName       string
	SpecName           string
	AcceptanceStandard string
}{
	{"青椒", "G0001", "蔬菜", "新鲜", "果面光洁、色泽青绿、无虫蛀、无腐烂、无异味"},
	{"西红柿", "G0002", "蔬菜", "新鲜", "果形完整、表皮红润、无裂口、无异味"},
	{"土豆", "G0003", "蔬菜", "新鲜散装", "块型饱满、表皮光滑、无芽眼发黑、无霉变"},
	{"大白菜", "G0004", "蔬菜", "新鲜散装", "菜叶紧实、无黄叶、无腐心、无异味"},
	{"胡萝卜", "G0005", "蔬菜", "新鲜散装", "色泽橙红、表皮光滑、无裂根、无腐烂"},

	{"苹果", "G0006", "水果", "新鲜", "果面光洁、无碰伤、无腐烂、口感清甜"},
	{"香蕉", "G0007", "水果", "新鲜", "果皮金黄、无黑斑、果梗新鲜、无异味"},
	{"雪梨", "G0008", "水果", "新鲜", "果形端正、果皮光洁、无霉烂、口感清脆"},
	{"西瓜", "G0009", "水果", "新鲜散装", "瓜形端正、瓜皮纹路清晰、无裂纹、无异味"},

	{"猪前腿肉", "G0010", "肉、禽、蛋", "新鲜", "肌肉紧实、色泽鲜红、脂肪洁白、无异味"},
	{"鸡腿", "G0011", "肉、禽、蛋", "新鲜不杀", "表皮干净、肉质紧实、无淤血、无异味"},
	{"鸡蛋", "G0012", "肉、禽、蛋", "新鲜散装", "壳面洁净、无破损、晃动无声、无异味"},
	{"牛腩", "G0013", "肉、禽、蛋", "新鲜", "肉色鲜红、纹理分明、无粘手、无异味"},
	{"鸭肉", "G0014", "肉、禽、蛋", "新鲜不杀", "表皮光洁、脂肪均匀、无淤血、无异味"},

	{"草鱼", "G0015", "水产海鲜", "新鲜不杀", "鱼眼饱满、鳞片完整、鱼鳃鲜红、无异味"},
	{"鲫鱼", "G0016", "水产海鲜", "新鲜不杀", "鱼体完整、鳞片紧密、腹部无膨胀、无异味"},
	{"罗非鱼", "G0017", "水产海鲜", "新鲜不杀", "鱼体完整、鱼腹紧实、鱼鳃鲜红、无异味"},
	{"基围虾", "G0018", "水产海鲜", "新鲜", "虾体透明、虾壳紧实、虾头完整、无异味"},

	{"纯牛奶", "G0019", "牛奶饮品", "250g", "包装完好、保质期内、乳香纯正"},
	{"酸奶", "G0020", "牛奶饮品", "180g", "低温保存、包装完好、质地均匀"},

	{"东北大米", "G0021", "干货调料", "500g", "米粒饱满、无虫蛀、无霉味"},
	{"高筋面粉", "G0022", "干货调料", "500g", "粉质洁白、无结块、无异味"},
	{"花生油", "G0023", "干货调料", "1.25L", "色泽金黄、无悬浮物、气味纯正"},
	{"生抽酱油", "G0024", "干货调料", "500g", "色泽红亮、口味鲜美、包装完好"},
	{"食用盐", "G0025", "干货调料", "500g", "晶粒洁白、干燥无结块、包装完好"},

	{"南豆腐", "G0026", "冻品、豆制品", "新鲜散装", "成型完整、色泽乳白、无酸臭味"},
	{"卤水豆干", "G0027", "冻品、豆制品", "新鲜散装", "块型完整、表面干爽、无霉斑"},
	{"冻鸡翅中", "G0028", "冻品、豆制品", "700g", "个体均匀、肉质紧实、无冰霜异常"},
	{"速冻虾仁", "G0029", "冻品、豆制品", "700g", "虾仁饱满、色泽自然、无异味"},
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
			Code:               item.Code,
			OrgID:              DefaultOrgID,
			SpecID:             specIDs[item.SpecName],
			CategoryID:         categoryIDs[item.CategoryName],
			AcceptanceStandard: toPtr(item.AcceptanceStandard),
			Sort:               0,
		}
		if err := upsertGoods(ctx, gdb, row); err != nil {
			logger.L().Error("Failed to seed goods", zap.Error(err),
				zap.String("name", item.Name),
				zap.String("code", item.Code),
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
