#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
市场价格数据导入脚本
用于将Excel表格中的市场价格数据导入到数据库
"""

import pandas as pd
import pymysql
import uuid
from datetime import datetime, date
from decimal import Decimal
from typing import Dict, List, Optional
import argparse
import sys

class MarketPriceImporter:
    """市场价格数据导入器"""
    
    def __init__(self, host: str, port: int, user: str, password: str, database: str):
        """初始化数据库连接"""
        self.conn = pymysql.connect(
            host=host,
            port=port,
            user=user,
            password=password,
            database=database,
            charset='utf8mb4'
        )
        self.cursor = self.conn.cursor()
        
    def __enter__(self):
        return self
        
    def __exit__(self, exc_type, exc_val, exc_tb):
        """关闭数据库连接"""
        if self.cursor:
            self.cursor.close()
        if self.conn:
            self.conn.close()
    
    def get_or_create_org(self, org_name: str = "都匀市") -> str:
        """获取或创建组织"""
        # 查询是否存在
        self.cursor.execute("SELECT id FROM base_org WHERE name = %s AND is_deleted = 0", (org_name,))
        result = self.cursor.fetchone()
        if result:
            return result[0]
        
        # 创建新组织（根组织，parent_id指向自己）
        org_id = str(uuid.uuid4())
        self.cursor.execute("""
            INSERT INTO base_org (id, name, code, sort, parent_id, description, is_deleted)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
        """, (org_id, org_name, org_name.upper(), 0, org_id, f"{org_name}组织", 0))
        self.conn.commit()
        return org_id
    
    def get_or_create_category(self, category_name: str, org_id: str) -> str:
        """获取或创建品类"""
        self.cursor.execute("""
            SELECT id FROM base_category 
            WHERE name = %s AND org_id = %s AND is_deleted = 0
        """, (category_name, org_id))
        result = self.cursor.fetchone()
        if result:
            return result[0]
        
        category_id = str(uuid.uuid4())
        self.cursor.execute("""
            INSERT INTO base_category (id, name, org_id, code, sort, is_deleted)
            VALUES (%s, %s, %s, %s, %s, %s)
        """, (category_id, category_name, org_id, f"CAT_{category_name}", 0, 0))
        self.conn.commit()
        return category_id
    
    def get_or_create_spec(self, spec_name: str) -> str:
        """获取或创建规格"""
        self.cursor.execute("""
            SELECT id FROM base_spec WHERE name = %s AND is_deleted = 0
        """, (spec_name,))
        result = self.cursor.fetchone()
        if result:
            return result[0]
        
        spec_id = str(uuid.uuid4())
        self.cursor.execute("""
            INSERT INTO base_spec (id, name, code, sort, is_deleted)
            VALUES (%s, %s, %s, %s, %s)
        """, (spec_id, spec_name, f"SPEC_{spec_name}", 0, 0))
        self.conn.commit()
        return spec_id
    
    def get_or_create_unit(self, unit_name: str) -> str:
        """获取或创建单位"""
        self.cursor.execute("""
            SELECT id FROM base_unit WHERE name = %s AND is_deleted = 0
        """, (unit_name,))
        result = self.cursor.fetchone()
        if result:
            return result[0]
        
        unit_id = str(uuid.uuid4())
        self.cursor.execute("""
            INSERT INTO base_unit (id, name, code, sort, is_deleted)
            VALUES (%s, %s, %s, %s, %s)
        """, (unit_id, unit_name, f"UNIT_{unit_name}", 0, 0))
        self.conn.commit()
        return unit_id
    
    def get_or_create_goods(self, goods_name: str, spec_id: str, unit_id: str, 
                           category_id: str, org_id: str) -> str:
        """获取或创建商品"""
        self.cursor.execute("""
            SELECT id FROM base_goods 
            WHERE name = %s AND spec_id = %s AND unit_id = %s 
              AND org_id = %s AND is_deleted = 0
        """, (goods_name, spec_id, unit_id, org_id))
        result = self.cursor.fetchone()
        if result:
            return result[0]
        
        goods_id = str(uuid.uuid4())
        goods_code = f"SKU_{goods_name}_{datetime.now().strftime('%Y%m%d%H%M%S')}"
        self.cursor.execute("""
            INSERT INTO base_goods (id, name, code, sort, spec_id, unit_id, 
                                   category_id, org_id, is_deleted)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
        """, (goods_id, goods_name, goods_code, 0, spec_id, unit_id, 
              category_id, org_id, 0))
        self.conn.commit()
        return goods_id
    
    def get_or_create_market(self, market_name: str, market_type: int, org_id: str) -> str:
        """获取或创建市场"""
        self.cursor.execute("""
            SELECT id FROM base_market 
            WHERE name = %s AND org_id = %s AND is_deleted = 0
        """, (market_name, org_id))
        result = self.cursor.fetchone()
        if result:
            return result[0]
        
        market_id = str(uuid.uuid4())
        self.cursor.execute("""
            INSERT INTO base_market (id, name, code, market_type, org_id, is_deleted)
            VALUES (%s, %s, %s, %s, %s, %s)
        """, (market_id, market_name, f"MKT_{market_name}", market_type, org_id, 0))
        self.conn.commit()
        return market_id
    
    def get_or_create_period(self, period_name: str, start_date: date, 
                            end_date: date, period_type: int, org_id: str) -> str:
        """获取或创建价格时期"""
        self.cursor.execute("""
            SELECT id FROM base_price_period 
            WHERE name = %s AND org_id = %s AND is_deleted = 0
        """, (period_name, org_id))
        result = self.cursor.fetchone()
        if result:
            return result[0]
        
        period_id = str(uuid.uuid4())
        period_code = start_date.strftime('%Y-%m-%d')
        self.cursor.execute("""
            INSERT INTO base_price_period (id, name, code, start_date, end_date, 
                                          period_type, org_id, is_deleted)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
        """, (period_id, period_name, period_code, start_date, end_date, 
              period_type, org_id, 0))
        self.conn.commit()
        return period_id
    
    def get_or_create_supplier(self, supplier_name: str, float_ratio: Decimal, org_id: str) -> str:
        """获取或创建供应商"""
        self.cursor.execute("""
            SELECT id FROM supplier 
            WHERE name = %s AND org_id = %s AND is_deleted = 0
        """, (supplier_name, org_id))
        result = self.cursor.fetchone()
        if result:
            return result[0]
        
        supplier_id = str(uuid.uuid4())
        self.cursor.execute("""
            INSERT INTO supplier (id, name, code, sort, status, description, 
                                 float_ratio, org_id, is_deleted)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
        """, (supplier_id, supplier_name, f"SUP_{supplier_name}", 0, 1, 
              f"{supplier_name}供应商", float_ratio, org_id, 0))
        self.conn.commit()
        return supplier_id
    
    def insert_market_price(self, goods_id: str, market_id: str, period_id: str,
                           price: Decimal, price_type: int, org_id: str):
        """插入市场价格"""
        price_id = str(uuid.uuid4())
        try:
            self.cursor.execute("""
                INSERT INTO base_market_price (id, goods_id, market_id, period_id,
                                              price, price_type, org_id, is_deleted)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                ON DUPLICATE KEY UPDATE 
                    price = VALUES(price),
                    updated_at = CURRENT_TIMESTAMP
            """, (price_id, goods_id, market_id, period_id, price, price_type, org_id, 0))
            self.conn.commit()
        except Exception as e:
            print(f"插入市场价格失败: {e}")
            self.conn.rollback()
    
    def insert_supplier_price(self, goods_id: str, supplier_id: str, period_id: str,
                             reference_price: Decimal, float_ratio: Decimal, org_id: str):
        """插入供应商结算价格"""
        price_id = str(uuid.uuid4())
        try:
            self.cursor.execute("""
                INSERT INTO base_supplier_price (id, goods_id, supplier_id, period_id,
                                                reference_price, float_ratio, org_id, is_deleted)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                ON DUPLICATE KEY UPDATE 
                    reference_price = VALUES(reference_price),
                    float_ratio = VALUES(float_ratio),
                    updated_at = CURRENT_TIMESTAMP
            """, (price_id, goods_id, supplier_id, period_id, reference_price, 
                  float_ratio, org_id, 0))
            self.conn.commit()
        except Exception as e:
            print(f"插入供应商价格失败: {e}")
            self.conn.rollback()
    
    def import_from_excel(self, excel_file: str, sheet_name: str, 
                         category_name: str, period_name: str,
                         start_date: date, end_date: date):
        """从Excel导入数据"""
        print(f"开始导入: {excel_file} - {sheet_name}")
        
        # 读取Excel
        try:
            df = pd.read_excel(excel_file, sheet_name=sheet_name, header=1)
        except Exception as e:
            print(f"读取Excel失败: {e}")
            return
        
        # 获取或创建基础数据
        org_id = self.get_or_create_org("都匀市")
        category_id = self.get_or_create_category(category_name, org_id)
        period_id = self.get_or_create_period(period_name, start_date, end_date, 1, org_id)
        
        # 创建市场
        markets = {
            '发改委指导价': (self.get_or_create_market('发改委', 1, org_id), 2),  # 价格类型：指导价
            '富万家超市': (self.get_or_create_market('富万家超市', 2, org_id), 1),  # 价格类型：市场价
            '育英巷菜市场': (self.get_or_create_market('育英巷菜市场', 3, org_id), 1),
            '大润发': (self.get_or_create_market('大润发', 2, org_id), 1),
            '上月均价': (None, 3),  # 价格类型：上月均价
            '本期均价': (None, 4),  # 价格类型：本期均价
        }
        
        # 创建供应商（用于结算价）
        supplier_hupu = self.get_or_create_supplier('胡埗', Decimal('0.88'), org_id)  # 下浮12%
        supplier_huanghai = self.get_or_create_supplier('黄海', Decimal('0.86'), org_id)  # 下浮14%
        
        # 列名映射
        column_mapping = {
            '品名': 'goods_name',
            '规格标准': 'spec_name',
            '单位': 'unit_name',
            '发改委指导价': 'guide_price',
            '富万家超市': 'fuwanjia_price',
            '育英巷菜市场': 'yuyingxiang_price',
            '大润发': 'darunfa_price',
            '上月均价': 'last_month_avg',
            '本期均价': 'current_avg',
            '胡埗本期结算价(下浮12%)': 'hupu_price',
            '黄海本期结算价(下浮14%)': 'huanghai_price',
        }
        
        # 遍历每一行
        for idx, row in df.iterrows():
            try:
                goods_name = str(row.get('品名', '')).strip()
                if not goods_name or goods_name == 'nan':
                    continue
                
                spec_name = str(row.get('规格标准', '新鲜')).strip()
                unit_name = str(row.get('单位', '斤')).strip()
                
                # 创建基础数据
                spec_id = self.get_or_create_spec(spec_name)
                unit_id = self.get_or_create_unit(unit_name)
                goods_id = self.get_or_create_goods(goods_name, spec_id, unit_id, 
                                                   category_id, org_id)
                
                print(f"处理商品: {goods_name} ({spec_name}/{unit_name})")
                
                # 插入各市场价格
                for col_name, (market_id, price_type) in markets.items():
                    price_value = row.get(col_name)
                    if pd.notna(price_value) and price_value != '':
                        try:
                            price = Decimal(str(price_value))
                            if market_id:
                                # 插入市场价格
                                self.insert_market_price(goods_id, market_id, period_id,
                                                        price, price_type, org_id)
                            else:
                                # 插入均价（不关联具体市场）
                                # 创建一个虚拟市场用于存储均价
                                avg_market_id = self.get_or_create_market(
                                    col_name, 5, org_id)  # 类型5=其他
                                self.insert_market_price(goods_id, avg_market_id, period_id,
                                                        price, price_type, org_id)
                        except (ValueError, TypeError) as e:
                            print(f"  警告: {col_name} 价格格式错误: {price_value}")
                
                # 插入供应商结算价
                current_avg = row.get('本期均价')
                if pd.notna(current_avg) and current_avg != '':
                    try:
                        reference_price = Decimal(str(current_avg))
                        # 胡埗供应商
                        self.insert_supplier_price(goods_id, supplier_hupu, period_id,
                                                  reference_price, Decimal('0.88'), org_id)
                        # 黄海供应商
                        self.insert_supplier_price(goods_id, supplier_huanghai, period_id,
                                                  reference_price, Decimal('0.86'), org_id)
                    except (ValueError, TypeError) as e:
                        print(f"  警告: 本期均价格式错误: {current_avg}")
                
            except Exception as e:
                print(f"处理第 {idx} 行时出错: {e}")
                continue
        
        print(f"完成导入: {sheet_name}\n")


def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='导入市场价格数据')
    parser.add_argument('--host', default='localhost', help='数据库主机')
    parser.add_argument('--port', type=int, default=3306, help='数据库端口')
    parser.add_argument('--user', default='food_user', help='数据库用户名')
    parser.add_argument('--password', default='StrongPassw0rd!', help='数据库密码')
    parser.add_argument('--database', default='main', help='数据库名')
    parser.add_argument('--file', required=True, help='Excel文件路径')
    parser.add_argument('--period', default='2025年9月上旬', help='价格时期名称')
    parser.add_argument('--start-date', default='2025-09-01', help='开始日期(YYYY-MM-DD)')
    parser.add_argument('--end-date', default='2025-09-10', help='结束日期(YYYY-MM-DD)')
    
    args = parser.parse_args()
    
    # 解析日期
    start_date = datetime.strptime(args.start_date, '%Y-%m-%d').date()
    end_date = datetime.strptime(args.end_date, '%Y-%m-%d').date()
    
    # 导入数据
    with MarketPriceImporter(args.host, args.port, args.user, 
                            args.password, args.database) as importer:
        # 可以导入多个sheet，每个sheet对应不同的品类
        sheets = [
            ('蔬菜类', '蔬菜类'),
            ('水产类', '水产类'),
            ('水果类', '水果类'),
        ]
        
        for sheet_name, category_name in sheets:
            try:
                importer.import_from_excel(
                    args.file, 
                    sheet_name, 
                    category_name,
                    args.period,
                    start_date,
                    end_date
                )
            except Exception as e:
                print(f"导入 {sheet_name} 失败: {e}")
                continue
    
    print("所有数据导入完成！")


if __name__ == '__main__':
    main()
