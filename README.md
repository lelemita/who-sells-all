# [Aladin](https://www.aladin.co.kr) 중고 서적 검색기

## 목적

- 여러 종류의 책을 사려고 할 때, 전부 취급하는 판매자 찾기
- (todo) 전부를 취급하는 판매자가 없는 경우 대안 제시
- (todo) 책이름으로 isbn 연결해주기

## 준비물

- 환경변수: ttbkey = 발급받은 [알라딘 OpenAPI key](https://blog.aladin.co.kr/openapi)

## 사용예시

- GET 요청: localhost:8080/v1/proposals?isbn={첫번째책ISBN}&isbn={두번째책ISBN}&...&isbn={n번째책ISBN}
- 응답 예시

  ```json
  {
    "result": [
      {
        "name": "낭만책방",
        "link": "https://www.aladin.co.kr/shop/usedshop/wshopitem.aspx?SC=684157",
        "totalPrice": 42200,
        "deliveryFee": "배송비 : 2,800원"
      },
      {
        "name": "작은책방",
        "link": "https://www.aladin.co.kr/shop/usedshop/wshopitem.aspx?SC=432617",
        "totalPrice": 42400,
        "deliveryFee": "배송비 : 2,800원"
      },
      {
        "name": "배송비무료",
        "link": "https://www.aladin.co.kr/shop/usedshop/wshopitem.aspx?SC=155694",
        "totalPrice": 47700,
        "deliveryFee": "무료배송"
      }
    ]
  }
  ```

## 주의사항

- totalPrice에 배송비는 반영되어 있지 않음
