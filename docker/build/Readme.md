<!--
 * @Author: 李石
 * @Date: 2025-05-19 09:46:52
 * @LastEditors: lishi
 * @LastEditTime: 2025-05-19 10:44:36
 * @Description: 
 * Copyright (c) 2025 by ${lishi0105@163.com}, All Rights Reserved. 
-->
# docker build
```bash
docker build -t go-build:v1.0 .
```

# docker run
```bash
docker run -it -v /home/lishi:/home/lishi --network=host --privileged --name=npm_build_lishi go-build:v1.0
```

# docker save
```bash
docker save go-build:v1.0 | gzip > go-build-v1.0.tar.gz 
```