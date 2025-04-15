我會提供一篇新聞,包含標題與內容,請你依照需求進行評分 

# 指標
標題與內容會有不同指標 , 指標與評分細節如下  
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

# 評分
1. 每一個指標你會給我一個0-5分的評分(score) , 指標下的項目需要檢查並斟酌扣分 , 分數越高代表越不包含這些負面指標
2. 每一個指標你會給我30字以下的評語(reason) 
3. 再根據這個指標分數加總平均到小數點下一位 , 得到總分(score) , 滿分5.0 , 並給一個整體評語(reason) 

# 格式
你會給我一個json格式,如下

```json
{
  "titleAnalytics":{
    "score":1.0,
    "reason":"a reason",
    "metricList":[
      {
        "metricKey":"accuracy",
        "score":1.0,
        "reason":"accuracy reason"
      }
    ]
  },
  "contentAnalytics":{
    "score":1.0,
    "reason":"a reason",
    "metricList":[
      {
        "metricKey":"accuracy",
        "score":1.0,
        "reason":"accuracy reason"
      }
    ]
  }
}
```