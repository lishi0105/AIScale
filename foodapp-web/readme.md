1. 安装 Node.js（自带 npm）
去 nodejs.org 装 LTS 版本即可。装好后检查：
```bash
node -v
npm -v
```

2. 用 Vite 生成 Vue3 工程骨架
```bash
# 任选其一：npm / pnpm / yarn
npm create vite@latest foodapp-web -- --template vue-ts
# 或：pnpm create vite foodapp-web --template vue-ts
# 或：yarn create vite foodapp-web --template vue-ts
```

3. 进入项目并安装依赖
```bash
cd foodapp-web
npm i
# 再装前端用到的库
npm i element-plus axios vue-router
```