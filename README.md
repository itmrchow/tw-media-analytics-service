# tw-media-analytics

分析新聞標題,內容 , 並給予評分

# arch

## Package arch

```
domain <- Service modules are independently executable
infra <- init function
main.go <- code init
```

## domain

## module

### di

### api_router

### logger
zerolog

### Observability
https://opentelemetry.io/docs/concepts/observability-primer/

### orm
gorm

### ai_model
AI model 操作

### spider
爬蟲, 使用go colly爬取新聞
### cron

### news
新聞資源操作

### queue

## 流程
cron 觸發爬蟲 job -> 爬蟲抓取新聞 -> 抓取新聞寫入資料庫 -> 新聞標題傳給AI model進行評分 -> 評分結果寫入資料庫



# 分析目標 target
- [ ] 電視新聞
  - [ ] 台視新聞台
  - [ ] 中視新聞台
  - [ ] 華視新聞資訊台
  - [ ] 民視新聞台
  - [ ] TVBS新聞台
  - [ ] 東森新聞台
  - [x] 三立新聞台
  - [x] 中天新聞台
  - [ ] 年代新聞台
  - [ ] 壹電視新聞台
  - [ ] 非凡新聞台
  - [ ] 寰宇新聞台
- [ ] 報社
  - [ ] 自由時報
  - [ ] 聯合報
  - [ ] 經濟日報
- [ ] 新興網路媒體
  - [ ] PeoPo公民新聞
  - [ ] ETtoday新聞雲
  - [ ] NOWnews今日新聞

# 評分方式
- 5-4 分：高品質新聞，值得推薦。
- 4-3 分：一般新聞，可供參考。
- 3-2 分：不推薦新聞，存在較多問題。
- 2-1 分：極不推薦新聞，品質低劣。
- 1-0 分：垃圾新聞，毫無價值。

## 分析面向指標
- 標題 
  - 準確性(accuracy)
    - 檢查包含虛假與錯誤、失實、捏造
    - 檢查包含標題與事實不符或內容不符
  - 清晰性(clarity)
    - 包含模糊與隱晦、隱瞞、誤導 
    - 包含「震驚」、「驚爆」、「絕對」、「史上最」
    - 包含暗示語氣
  - 客觀性(objectivity)
    - 包含主觀與偏頗、偏見、有失公正
    - 包含情緒化字眼 
  - 相關性(relevance)
    - 檢查與內容相關
  - 吸引力(attractiveness)
    - 無過度且欺騙的吸引力、聳動、欺騙
    - 故意留下懸念，如:竟然…
    - 包含「曝」
- 內容
  - 準確性(accuracy):
    - 虛假性與錯誤性、失實、捏造 
  - 客觀性(objectivity)
    - 主觀性與偏頗性、偏見、有失公正 
  - 即時性(timeliness)
    - 滯後性與過時性、過期、落後 
  - 重要性(importance)
    - 瑣碎性與無關性、無關緊要、不重要 
  - 呈現性(presentation)
    - 混亂與粗糙。粗俗、不專業
    - 新聞來源是節目 , 內容是節目對話




# log format

```
{
  "domain"
  "traceId"
  
}
```