
# coding=utf-8
import json
import requests
from lxml import etree

result = dict()

def crawl():
    resp = requests.get("https://uutool.cn/info-nation/", timeout=120)
    content = resp.text
    html = etree.HTML(content)
    tr_list = html.xpath("//tbody//tr")
    print(len(tr_list))
    for tr in tr_list:
        chinese_name = tr.xpath("./td[6]/text()")[0]
        # chinese_name = chinese_name.decode()
        chinese_name = chinese_name.strip()
        eng_iso2 = tr.xpath("./td[1]/text()")[0]
        eng_iso3 = tr.xpath("./td[2]/text()")[0]
        english_name = tr.xpath("./td[5]/text()")[0]
        result[chinese_name] = {"iso2": eng_iso2, "iso3": eng_iso3, "eng": english_name}
    if len(result) > 0:
        with open("country_names.json", "w", encoding="utf-8") as f:
            json.dump(result, f, indent=4, ensure_ascii=False)
    
if __name__ == '__main__':
    crawl()
