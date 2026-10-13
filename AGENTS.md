# gortsplib (JDONE 포크) — 에이전트 작업 규칙

이 저장소에서 일하는 AI 코딩 에이전트(Claude Code·Codex·Cursor 등)와 사람이 함께 따르는 규칙이다.

- Claude Code 는 `CLAUDE.md` 가 이 파일을 가져오고, Codex·Cursor 는 이 파일을 바로 읽는다.
- Cursor Bugbot 리뷰 지침은 `.cursor/BUGBOT.md`, 커밋·PR·화면 캡처 절차 스킬은 `.agents/skills/`(Claude Code 는 같은 내용을 가리키는 `.claude/skills/` 링크)에 있다.
- 「작업 흐름」·「커밋 메시지」·「PR」은 JDONE 저장소 공통 규칙이다. 바꿀 때는 다른 저장소도 함께 맞춘다.

## 기본 원칙

- 사용자에게 보이는 글(답변·커밋·PR·리뷰 답글)은 한국어로 쓴다. 코드·식별자·명령·로그 원문은 그대로 둔다.
- 새 관례를 만들지 말고 주변 코드 스타일을 따른다. 고치기 전에 같은 파일의 import 블록과 정의 순서를 먼저 본다.
- 쓰지 않게 된 코드·헬퍼·설정은 남기지 말고 지운다. 설정 키를 바꾸거나 없앨 때 옛 키 별칭·폐기 키 호환을 두지 않는다.
- 비밀번호·토큰·키·고객 실데이터를 코드·커밋·PR·로그에 남기지 않는다.
- 커밋·push·PR·머지는 사용자가 시킨 단계까지만 한다. `fix/digest-qop-auth` 에 직접 push 하지 않는다. 서버 배포는 사용자가 명시적으로 시킬 때만 한다.
- 커밋·PR 에 `Co-Authored-By:`·`Generated with …` 같은 도구 서명 줄을 넣지 않는다. 도구 기본 설정보다 이 규칙이 우선이다.
- 저장소에 올리는 문서(README·docs·주석)에는 작업 환경·남은 일·작업 경과 같은 작업 메모를 쓰지 않는다. 문서는 공적이고 간결한 톤으로 쓴다.
- 설정 파일·템플릿·예시 파일에 항목을 더할 때 주석은 같은 파일의 다른 항목과 같은 형식으로 쓴다.
- 서버에 설치·배포한 파일에 작업 흔적(날짜·사유·추가 표시 주석, AI 관련 표현, 임시 작업 폴더)을 남기지 않는다. 임시로 만든 SSH 키·비밀 파일은 작업이 끝나면 지운다.
- 코드 주석·doc·오류 문구는 upstream 처럼 영어로 쓴다. 한국어는 커밋·PR·리뷰 답글에만 쓴다.

## 저장소 개요

- bluenviron/gortsplib(RTSP 클라이언트·서버 라이브러리) 포크. 모듈 경로는 upstream 그대로 `github.com/bluenviron/gortsplib/v5`, `go 1.25`, pion rtp·rtcp·srtp·sdp, mediacommon v2.
- JDONE 변경은 RTSP Digest `qop="auth"`(RFC 7616) 응답 계산과 Opaque 전달 패치(`555a3030`: `pkg/auth/sender.go`, `pkg/headers/authenticate.go`·`authorization.go`, 테스트 `pkg/auth/sender_qop_test.go`)다.
- 기준은 upstream `v5.6.1`(`b46fd179`). **기본 브랜치가 `fix/digest-qop-auth` 이고 origin 에 `main` 은 없다.** `upstream` 원격(bluenviron)이 설정돼 있다.
- 쓰는 곳: ai-media-one·vms-gateway 의 `go.mod` `replace github.com/bluenviron/gortsplib/v5 => github.com/JDONE-Research/gortsplib/v5 <pseudo-version>`.

## 구조

- 루트 패키지 `gortsplib`: `client*.go`, `server*.go`(Client·Server·ServerConn·ServerSession·ServerStream). 파일 이름은 `<역할>_<세부>.go`.
- `pkg/`: 공개 하위 패키지(`auth`·`base`·`headers`·`format`·`description`·`sdp`·`liberrors`·`rtpreceiver`·`rtpsender`·`ntp` …)
- `internal/`: 내부 유틸, `internal/teste2e`(`-tags enable_e2e_tests`, 도커). `examples/`: 사용 예제(일부 cgo ffmpeg).
- `scripts/*.mk`, `.golangci.yml`(ai-media-one 과 같은 린터), `.github/workflows/{lint,test}.yml`, `.github/dependabot.yml`
- 새 코드: upstream 구조를 그대로 두고 해당 패키지 파일을 최소로 고친다. 테스트는 새 파일 `<대상>_<주제>_test.go`(예: `sender_qop_test.go`).

## 명령

```sh
make test-nodocker      # examples 빌드 + ./internal/... + ./pkg/... + 루트, 64비트면 -race
go test ./pkg/auth -run TestSenderDigestQOPAuth -count=1   # 단일
make test               # 도커 amd64 (ffmpeg-dev 포함)
make test-32            # 도커 i386
make test-e2e           # 도커 (또는 make test-e2e-nodocker)
make lint               # golangci-lint (도커)
go mod tidy -diff       # CI go_mod job
make format             # gofumpt (도커)
git fetch upstream && git log --oneline HEAD..upstream/main   # upstream 과의 차이
```

- **CI 는 `main` 브랜치에만 걸려 있어 이 저장소의 PR 에서는 돌지 않는다.** PR 전에 `make test-nodocker`·`make lint`·`go mod tidy -diff` 를 반드시 직접 돌린다.

### 바뀐 범위 테스트 (작업 중)

브랜치를 딴 뒤 바꾼 파일(커밋한 것·안 한 것·새 파일)과 관련된 테스트만 돈다. 관련 테스트가 없으면 아무것도 돌지 않는다. (bash·zsh, 맥·리눅스 공통)

```bash
{ git diff --name-only --diff-filter=d $(git merge-base origin/HEAD HEAD); git ls-files -o --exclude-standard; } | sort -u | grep '\.go$' | xargs -r -n1 dirname | sort -u | sed 's#^#./#' | xargs -r go test -race
{ git diff --name-only --diff-filter=d $(git merge-base origin/HEAD HEAD); git ls-files -o --exclude-standard; } | sort -u | grep '\.go$' | xargs -r gofmt -l   # 출력이 없어야 한다
```

- 바꾼 `.go` 파일이 있는 패키지만 시험한다. 공개 함수·타입·인터페이스를 바꿨으면 그것을 쓰는 패키지도 더한다(`grep -rl '"<모듈 경로>/<패키지>"' --include='*.go' .`).
- 실패한 것만 다시: `go test -run '<시험 이름>' ./<패키지>`

### 커밋 직전 전체 검사 (한 번)

```bash
make test-nodocker
make lint
go mod tidy -diff
```

- 이 저장소의 PR 에는 CI 가 돌지 않으니 결과를 PR 「확인」에 적는다.

## 코드 스타일

### 원칙

- upstream(aler9 / Alessandro Ros) 스타일을 그대로 따른다. JDONE 변경은 "패치"로 보고 diff 를 최소로 한다. 기존 줄을 지우거나 다듬지 않는다(주석 삭제·리포맷·이름 변경 금지).
- **코드 주석·doc·오류 문구는 upstream 처럼 영어로 쓴다.** 한국어는 커밋 메시지·PR 에만 쓴다.

### 포맷·린트

- ai-media-one 과 같은 `.golangci.yml`: gofmt·gofumpt·goimports, `lll` 120, `whitespace`, `revive`, govet `shadow` strict, `errorlint`, `modernize` 등.

### 선언 순서·빈 줄·import

- **비공개 최상위 헬퍼를 그 헬퍼를 쓰는 공개 함수·메서드보다 위에** 둔다(upstream 은 예외 없이 이렇다). 새 헬퍼는 파일 위쪽이나 같은 패키지의 헬퍼 파일(`verify.go` 의 `md5Hex` 옆)에 둔다.
- 메서드의 공개·비공개 순서 규칙은 없다. 구조체는 공개 설정 필드를 채운 뒤 초기화 메서드를 부르는 패턴이고, 이름은 패키지마다 다르다(`pkg/format/rtp*` 는 `Init() error`, `auth.Sender`·`ServerStream` 은 `Initialize() error`, `Client`·`Server` 는 `Start() error`).
- 최상위 선언 사이 빈 줄 한 줄(gofmt). 블록이 끝나고 다음 문장이 오면 빈 줄을 두는 편이다.
- import 는 고치는 파일의 기존 형태를 유지한다(표준 / 외부 / `github.com/bluenviron/gortsplib/v5/...` 세 묶음이 더 많고, 외부와 섞은 두 묶음도 있다). 새 파일은 같은 패키지 이웃 파일을 따른다.

### doc·주석·오류

- 공개 식별자는 전부 doc 주석, 이름으로 시작한다(`// Sender allows to send credentials.`). 비공개에는 거의 달지 않는다.
- 헤더 구조체 필드는 소문자 키 이름 한 줄 주석(`// qop`).
- 오류는 `fmt.Errorf` 평문이 대부분이고 감쌀 때 `%w`. 공개 에러 타입은 `pkg/liberrors`. 메시지는 소문자 시작, 마침표 없음.

```go
// Sender allows to send credentials.
type Sender struct {
	WWWAuth base.HeaderValue
	User    string
	Pass    string

	authHeader *headers.Authenticate
}
```

```go
func hasAuthQOP(qop string) bool {
	for v := range strings.SplitSeq(qop, ",") {
		if strings.TrimSpace(v) == "auth" {
			return true
		}
	}
	return false
}
```

## 테스트

- testify `require`(assert 는 쓰지 않는다), 같은 패키지 안 테스트. Fuzz 테스트가 많다.
- 케이스는 패키지 수준 `var cases<대상> = []struct{ name string; ... }{...}` 를 두고 `for _, ca := range cases<대상> { t.Run(ca.name, ...) }` 로 돈다. 이 표를 Unmarshal·Marshal 테스트와 Fuzz 시드(`f.Add`)가 같이 쓴다. 한 테스트에서만 쓰는 표는 인라인 `[]struct`. 헬퍼(`mustParseURL`)는 패키지 안에서 재사용한다.
- 클라이언트·서버 테스트는 `net.Listen` 가짜 상대와 `base.Request`·`base.Response` 를 직접 주고받는다. e2e 는 `internal/teste2e`.
- 새 동작은 새 테스트 파일로 만들고, 기존 동작이 그대로인지 확인하는 케이스를 같이 넣는다(qop 없는 RFC 2069 응답 불변). 테스트 주석도 영어.

## 주의할 점

- 모듈 경로(`github.com/bluenviron/gortsplib/v5`)를 바꾸지 않는다. 쓰는 쪽의 replace 가 원래 경로를 전제로 한다.
- 이 저장소에 커밋이 머지되면 ai-media-one·vms-gateway 두 곳의 `go.mod` replace pseudo-version 을 새 커밋으로 올리고 `go mod tidy` 한다(각 저장소에서 별도 PR).
- upstream 최신을 받으려면 JDONE 패치를 새 기준 위로 옮겨야 한다. upstream 이 같은 파일(auth·headers)을 고쳤으면 충돌하니 테스트로 확인한다.
- dependabot 이 매일 PR 을 연다(base 는 기본 브랜치). 쓰는 쪽에 반영할 필요가 있는지 보고 머지한다.
- Digest 응답의 `nc` 가 늘 `00000001` 이고 `generateCnonce` 는 `rand.Read` 오류를 무시한다. 이 부분을 고칠 때는 RFC 7616 과 엄격한 서버 동작을 함께 본다.
- 기본 브랜치 HEAD 도 `make lint` 에 걸린다: JDONE 패치의 `hasAuthQOP` 가 `range strings.Split` 을 써서 modernize(stringsseq)에 걸리고, `generateCnonce`·`hasAuthQOP` 가 그것을 쓰는 `AddAuthorization` 아래에 있다. `pkg/auth/sender.go` 를 고치는 PR 에서는 두 헬퍼를 위로 옮기고 `strings.SplitSeq` 로 바꾼다.
- qop 지원은 클라이언트 송신 쪽만 있다. `headers.Authorization.Unmarshal` 은 qop·nc·cnonce 를 읽지 않고 `headers.Authenticate.Marshal` 은 qop 를 내보내지 않으며 `auth.Verify` 도 레거시 응답만 검증한다. 대칭 테스트(`casesAuthorization`)에 qop 케이스를 넣거나 서버가 qop 를 요구하게 하려면 이 부분부터 구현한다.

## 작업 흐름

개발 → 검토 → 커밋 직전 전체 테스트 → 커밋 → PR → 클라우드 리뷰 → 수정 → `fix/digest-qop-auth` 에 squash 머지. 검토와 전체 테스트는 매 수정마다 하지 않고 커밋 직전에 한 번씩만 한다. 깊은 버그 검토는 클라우드 리뷰가 맡는다. 절차는 스킬 `jdone-commit`·`jdone-pr` 에 있다.

1. **브랜치**: 최신 `fix/digest-qop-auth` 에서 `<타입>/<짧은-영어-설명>`(타입 `feat` `fix` `perf` `refactor` `style` `test` `docs` `chore` `revert`, 예: `fix/onvif-tls-compat`). `codex/`·`cursor/` 같은 도구 이름 접두어는 쓰지 않는다.
2. **개발**: 고치는 동안에는 「명령」의 **바뀐 범위 테스트**만 돌린다. 동작을 바꾸면 테스트를 더하거나 고친다.
3. **검토 (한 번)**: 기능이 다 되면 `git diff $(git merge-base origin/HEAD HEAD)` 를 한 번 읽고 「코드 스타일」과 `.cursor/BUGBOT.md` 「특히 볼 것」으로 점검해 고친다. 고친 곳은 바뀐 범위 테스트로 확인한다.
4. **전체 테스트 (커밋 직전 한 번)**: 「명령」의 커밋 직전 전체 검사를 돌린다(오래 걸리면 백그라운드로 돌리고 그동안 커밋 메시지를 쓴다). 실패하면 고친 뒤 실패한 것만 다시 돌리고, 고친 곳이 공용 코드일 때만 전체를 다시 돈다. 문서·주석만 바꾼 커밋은 생략한다.
5. **커밋**: 「커밋 메시지」 규칙대로, 한 커밋에 한 가지 목적. push 한 커밋은 amend·rebase·force push 하지 않고 새 커밋을 더한다.
6. **PR**: `git push -u origin HEAD` → `gh pr create --base fix/digest-qop-auth`(「PR」 규칙). PR 은 한 가지 목적으로 작게 만든다. 스택 PR(다른 PR 브랜치를 base 로) 은 만들지 않는다 — squash 머지라 아래 PR 이 머지되면 위 PR 이 닫힌다. 여러 PR 로 나눌 때는 앞 PR 이 머지된 `fix/digest-qop-auth` 기준으로 다음 브랜치를 만들고, 자동 머지될 수 있으니 올리는 순서에 유의한다.
7. **클라우드 리뷰**: PR 을 열거나 push 하면 Cursor 가 리뷰한다 — `Cursor Bugbot`(버그·인라인 지적), `Cursor Security Agent`(보안), `Cursor Approval Agent`(위험도 판정 → 자동 승인 또는 사람 리뷰어 배정). 이 저장소의 GitHub Actions 는 `main` 에만 걸려 있어 PR 에서 돌지 않는다. 로컬 `make test-nodocker`·`make lint` 결과를 PR 「확인」에 적는다. 끝날 때까지 기다린다(`gh pr checks <N>`).
8. **리뷰 검토·수정**: 지적마다 코드를 읽고 가능하면 테스트로 재현한다.
   - 맞으면 지적을 모두 고친 뒤 바뀐 범위 테스트 → 전체 검사(4) 한 번 → 커밋·push 하고, 스레드에 고친 커밋과 내용을 답한 뒤 resolve 한다.
   - 틀리면 근거(코드 위치·조건·의도된 설계)를 답하고 resolve 한다. 답 없이 넘기지 않는다.
   - Bugbot Autofix 커밋도 검토하고, 동작이 바뀌었으면 테스트로 고정한다. 사람 리뷰어가 배정되면 승인을 기다린다.
9. **머지**: 체크가 통과하고 승인되면 `gh pr merge <N> --squash`(제목 `PR 제목 (#N)`, 본문은 GitHub 기본값). 머지한 뒤에는 ai-media-one·vms-gateway 의 `go.mod` replace 를 새 커밋으로 올리는 PR 을 각각 낸다.
   - Cursor 사용량 한도로 리뷰가 돌지 않으면(`usage limit reached`) 에이전트가 직접 깊게 검토한다: 영역별로 버그·보안·스타일을 보고, 지적마다 반박해 본 뒤 확인된 것만 고치고, 결과를 PR 코멘트로 남긴 다음 머지한다.
10. **정리**: `git switch fix/digest-qop-auth && git pull --ff-only` 후 머지된 브랜치(로컬·원격)를 지운다. 다른 워크트리에 체크아웃된 브랜치는 `gh pr merge --delete-branch` 가 워크트리까지 지우므로 쓰지 않는다. 머지 뒤 늦게 달린 Bugbot 지적은 후속 PR 로 고친다.

사용자가 작은 수정을 "PR 없이 fix/digest-qop-auth 에 바로" 라고 할 때만 `git merge --ff-only` 후 push 한다.

## 커밋 메시지

```text
<이모지> <타입>: <무엇이 어떻게 달라지는지 한 문장>

<배경 단락(선택): 요구사항 ID, 현상, 원인>

[영역]
- 파일 이름: 바꾼 내용
  (이유·근거·수치·버린 대안)
- 파일 이름1, 파일 이름2: 함께 바꾼 내용

[Test]
- 테스트 파일 이름: 무엇을 확인하는지
```

- `✨ Feat`: 새로운 기능이나 로직이 추가된 경우
- `🐞 Fix`: 오류나 잘못된 동작을 수정한 경우
- `🧹 Chore`: 작은 유지보수 작업이나 기타 잡다한 수정사항이 있는 경우
- `♻️ Refactor`: 기능 변경 없이 코드의 구조나 품질을 개선한 경우
- `📝 Docs`: 문서나 주석을 추가 또는 수정한 경우
- `🧪 Test`: 테스트를 추가하거나 수정한 경우
- `🚀 Perf`: 성능 개선을 위한 변경 사항이 있는 경우
- `⚙️ Config`: 설정 관련 파일이나 환경 설정을 변경한 경우
- `🚑️ Hotfix`: 긴급하게 문제를 수정한 경우
- `📦 Deploy`: 배포와 관련된 변경사항이 있는 경우
- `📁 Assets`: 정적 파일이나 리소스를 추가 또는 수정한 경우
- `📌 Package`: 패키지 설정 또는 종속성을 추가, 제거, 업데이트한 경우
- `🗄️ DB`: 데이터베이스 관련 설정이나 마이그레이션 작업을 수행한 경우
- `🔥 Remove`: 불필요한 코드나 파일을 삭제한 경우
- `🔀 Merge`: 다른 브랜치를 병합한 경우
- `⏪ Revert`: 이전 상태로 변경사항을 되돌린 경우 (제목 끝에 `(#N 되돌림)`)
- `🔒 Security`: 보안 관련된 수정이나 개선사항이 있는 경우
- `📊 Analytics`: 분석, 로깅 또는 추적 관련 변경사항이 있는 경우
- `🎨 Style`: 코드 스타일 변경 또는 서식 수정사항이 있는 경우
- `🚧 WIP`: 작업이 진행 중이며 임시 저장이 필요한 경우
- `🎉 Init`: 프로젝트를 처음 구성하거나 초기 설정을 추가한 경우

목록에 없는 이모지·타입(`🐛 Fix`·`⚡ Perf`·`💄 Style`·`🔧 Fix` 등)은 쓰지 않는다.

**제목**

- 한국어 한 문장, 현재형 "~한다"로 끝낸다 (…를 더한다, …지 않게 한다, …를 고친다). 마침표는 찍지 않는다.
- 코드 이름보다 동작·결과를 쓴다. 핵심 이유는 `—` 뒤에 붙이고, 설정 키는 괄호로 적는다 (예: `(s3.encryption)`).
- `(#N)` 은 쓰지 않는다. squash 머지 때 GitHub 가 붙인다.
- 금지: 영어 제목, `fix: …` 같은 소문자 Conventional Commits 형식, 이모지·타입 없는 제목.

**본문**

- 제목 다음 줄은 반드시 비운다 (없으면 git 이 본문 첫 단락까지 제목으로 읽는다).
- 영역 이름: `[pkg/<패키지>]`(예: `[pkg/auth]`·`[pkg/headers]`), 루트 패키지는 `[client]`·`[server]`, `[Test]`. 화면·기능 단위로 나누는 편이 읽기 쉬우면 한국어 영역 이름(예: `[카메라 설정]`, `[이벤트 이력]`)도 쓴다. `[Test]` 는 코드 영역 뒤, `[Docs]` 는 맨 끝.
- 항목은 `- 파일 이름: 내용`. 경로 없이 파일 이름만 쓰고(같은 이름이 여럿이면 구분되는 만큼 경로), 새 파일은 `(신규)`. 한 파일에 바꾼 것이 많으면 두 칸 들여 하위 목록.
- 이유·근거·수치·버린 대안은 다음 줄에 두 칸 들여 `(…)` 로 쓴다. 여러 줄이어도 된다.
- 항목은 개조식(…함, …음, …지 않음, 명사)으로 끝내고, 배경 단락만 평서문(~다)으로 쓴다.
- 이유는 diff·대화·실측으로 아는 것만 쓴다. 지어내지 않는다.
- 한두 줄짜리 작은 변경은 제목만 써도 된다.

예시:

```text
🐞 Fix: 조명·김을 화재로, 젖은 도로를 침수로 보던 VLM 오탐을 줄이고, 이전 판정 문장을 다음 분석이 베끼지 않게 한다

[Backend]
- analysis.py: 최근 이벤트 이력에서 장면 설명을 빼고 유형·심각도·지난 시간만 남김
  (최근 3건 설명이 다음 프롬프트에 들어가면 모델이 그 문장을 베껴 같은 오탐을 반복함. 시험 카메라 26건, 여차항 12건은 설명이 글자까지 같았음)
  (이력을 통째로 빼면 실제 불을 놓쳐, 유형·시각은 남김. 설명을 판정보다 먼저 쓰게 하면 실제 침수를 놓쳐 넣지 않음)
- seed_event_type_configs.sql: 기상재해에서 차선·연석·노면이 보이는 젖은 도로, 물웅덩이, 반사, 지면 그림자는 침수가 아님

[Test]
- test_vlm_analysis.py: 이력에 예전 설명 문장이 없고, 문장을 재사용하지 말라는 지시가 있는지
```

## PR

- 제목은 커밋 제목과 같은 형식이다. 커밋이 하나면 그 제목 그대로, 여럿이면 전체를 아우르는 한 문장.
- 본문은 한국어로 쓴다. 기능이면 `## 문제` 대신 `## 내용` 에 무엇을 왜 하는지 쓴다.

```markdown
## 문제

현상과 원인 (요구사항 ID 가 있으면 함께)

## 수정

- 바꾼 것 (동작 중심, 설정 키·API·파일 이름은 `코드`)

## 확인

- 돌린 테스트(명령·개수), 장비·서버에서 확인한 것, 확인하지 못한 것
```

- Cursor 가 본문 끝에 붙이는 요약 블록은 건드리지 않는다.
- 리뷰 답글도 한국어로 쓰고, 고쳤으면 커밋 해시와 고친 내용을 적는다.
