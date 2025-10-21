#!/bin/bash
# 市场价格管理系统快速启动脚本

set -e

echo "========================================"
echo "市场价格管理系统 - 快速部署"
echo "========================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 数据库配置
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_ROOT_USER="${DB_ROOT_USER:-root}"
DB_ROOT_PASS="${DB_ROOT_PASS}"
DB_USER="${DB_USER:-food_user}"
DB_PASS="${DB_PASS:-StrongPassw0rd!}"
DB_NAME="${DB_NAME:-main}"

# 检查MySQL是否已安装
check_mysql() {
    echo -e "${YELLOW}[1/6] 检查MySQL环境...${NC}"
    if ! command -v mysql &> /dev/null; then
        echo -e "${RED}错误: 未找到MySQL客户端，请先安装MySQL${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ MySQL客户端已安装${NC}"
}

# 创建数据库和用户
create_database() {
    echo -e "\n${YELLOW}[2/6] 创建数据库和用户...${NC}"
    if [ -z "$DB_ROOT_PASS" ]; then
        mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_ROOT_USER" < sql/00_db_users.sql
    else
        mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_ROOT_USER" -p"$DB_ROOT_PASS" < sql/00_db_users.sql
    fi
    echo -e "${GREEN}✓ 数据库和用户创建完成${NC}"
}

# 创建系统基础表
create_system_tables() {
    echo -e "\n${YELLOW}[3/6] 创建系统基础表...${NC}"
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < sql/01_base_sys.sql
    echo -e "${GREEN}✓ 系统基础表创建完成${NC}"
}

# 创建商品相关表
create_goods_tables() {
    echo -e "\n${YELLOW}[4/6] 创建商品相关表...${NC}"
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < sql/10_goods_domain.sql
    echo -e "${GREEN}✓ 商品相关表创建完成${NC}"
}

# 创建市场价格管理表
create_price_tables() {
    echo -e "\n${YELLOW}[5/6] 创建市场价格管理表...${NC}"
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < sql/11_market_price_system.sql
    echo -e "${GREEN}✓ 市场价格管理表创建完成${NC}"
}

# 插入示例数据（可选）
insert_sample_data() {
    echo -e "\n${YELLOW}[6/6] 是否插入示例数据？(y/n)${NC}"
    read -r answer
    if [ "$answer" = "y" ] || [ "$answer" = "Y" ]; then
        mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < sql/12_sample_data.sql
        echo -e "${GREEN}✓ 示例数据插入完成${NC}"
    else
        echo -e "${YELLOW}跳过示例数据插入${NC}"
    fi
}

# 验证安装
verify_installation() {
    echo -e "\n${YELLOW}验证安装...${NC}"
    TABLE_COUNT=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$DB_NAME'")
    echo -e "${GREEN}✓ 共创建 $TABLE_COUNT 张表${NC}"
    
    echo -e "\n${GREEN}========================================${NC}"
    echo -e "${GREEN}安装完成！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo -e "\n数据库信息:"
    echo -e "  主机: ${GREEN}$DB_HOST:$DB_PORT${NC}"
    echo -e "  数据库: ${GREEN}$DB_NAME${NC}"
    echo -e "  用户名: ${GREEN}$DB_USER${NC}"
    echo -e "\n下一步:"
    echo -e "  1. 查看使用说明: ${GREEN}cat sql/README_market_price.md${NC}"
    echo -e "  2. 安装Python依赖: ${GREEN}pip install -r sql/requirements.txt${NC}"
    echo -e "  3. 导入Excel数据: ${GREEN}python sql/import_market_prices.py --help${NC}"
    echo ""
}

# 主流程
main() {
    check_mysql
    create_database
    create_system_tables
    create_goods_tables
    create_price_tables
    insert_sample_data
    verify_installation
}

# 处理命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --host)
            DB_HOST="$2"
            shift 2
            ;;
        --port)
            DB_PORT="$2"
            shift 2
            ;;
        --root-user)
            DB_ROOT_USER="$2"
            shift 2
            ;;
        --root-pass)
            DB_ROOT_PASS="$2"
            shift 2
            ;;
        --help)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --host HOST          数据库主机 (默认: localhost)"
            echo "  --port PORT          数据库端口 (默认: 3306)"
            echo "  --root-user USER     root用户名 (默认: root)"
            echo "  --root-pass PASS     root密码"
            echo "  --help               显示此帮助信息"
            echo ""
            echo "示例:"
            echo "  $0 --host localhost --port 3306 --root-pass mypassword"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            echo "使用 --help 查看帮助"
            exit 1
            ;;
    esac
done

# 执行主流程
main
