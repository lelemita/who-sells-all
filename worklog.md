# WORK LOG

## 2026.0608.Monday.

### Fix Bug: 종종 결과가 안 나오는 현상
- 원인: 복수 ISBN 조회 시 Goroutine의 비동기 실행 순서와 채널 수신 순서가 일치하지 않아, 반환된 데이터가 어떤 ISBN의 결과인지 식별 불가했음.
- 위치: `getProposals()`에서 `go getForOneIsbn()`를 사용하는 부분
- 해결: 결과 전송 채널 구조체에 ISBN 정보를 포함하여, 결과와 ISBN을 명확히 매핑하도록 개선.
    - AS-IS: `chan Proposals`
    - TO-BE: `chan isbnResult` (ISBN과 Proposals를 포함한 구조체)



## TODOs
### search free server: cloudtype -> ?
    - cloudtype은 하루 한번씩 꺼짐
### log 관리
    - ctx 활용
    - 가독성? 색상? beautify?
    - type?: text -> json?
    - error 재현 가능하게 기록
    - 꺼질 때 알림?
### 에러 관리: slack?
### 광고달기
