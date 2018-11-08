# GoTuner 调律器

## 项目结构

+ 库：tuning（调音）
+ 框架：tuner（调律器）
+ 服务：server
+ Cli：client

## tuning 库 - 任务调度

概念解释：

+ Cantor   - 控制中心
+ Music    - 任务工厂
+ Track(s) - 任务（链）
+ Melody   - 任务负载
+ Clock    - 定时用

## tuner 框架 - 命令组整合

概念解释：

+ Tuner           - 调律器
+ CommandMelody   - 进程的封装
+ StreamFilter    - 流过滤器
+ Diverter        - 分流器
+ InputRectifier  - 输入整流器
+ OutputRectifier - 输出整流器

## TODO

+ [ ] tuning 优先级策略
+ [x] fix uuid
+ [ ] 协程池环境隔离、通用化
+ [ ] 协程池支持协奏（条件满足时同时启动）
+ [ ] 重奏的检查（Melody状态的保留与重奏冲突）
+ [ ] tuning - Track 允许给后继发 sign，Melody 实现自定义 sign 处理
+ [ ] 重构测试代码，细分公私方法对象变量
+ [ ] 截流器（用于暂停）
+ [ ] 添加进程池
+ [ ] 各部分使用不同的协程池
+ [ ] 处理 “子进程需要手动处理输入流关闭的情况” 
+ [ ] 处理 “子进程的输出流关闭前读取协程未读完的情况”
+ [ ] tuning 中的各概念支持携带自定义数据
+ [ ] tuner 重构，支持重奏
