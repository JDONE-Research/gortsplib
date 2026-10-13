---
name: jdone-pr
description: 커밋된 브랜치를 push 하고 JDONE 형식 PR 을 연 뒤, Cursor 클라우드 리뷰(Bugbot·Security Agent·Approval Agent)를 기다려 지적을 검토·수정하고 기본 브랜치에 squash 머지한 다음 브랜치를 정리한다. PR 생성, 리뷰 대응, 머지를 요청받았을 때 쓴다.
---

# JDONE PR → 클라우드 리뷰 → 수정 → 머지

규칙의 원문은 저장소 루트 `AGENTS.md` 의 「작업 흐름」·「PR」이다. 사용자가 맡긴 단계까지만 한다 (예: "PR 까지만" 이면 3단계에서 PR 주소를 보고하고 멈춘다).

```bash
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
BASE=$(gh repo view --json defaultBranchRef -q .defaultBranchRef.name)
```

## 1. 준비

- 현재 브랜치가 `$BASE` 가 아니고 `git log origin/$BASE..HEAD` 에 커밋이 있는지, 작업 트리가 깨끗한지 확인한다.
- 커밋 메시지가 `AGENTS.md` 형식(`<이모지> <타입>: 한국어 한 문장`)인지 `git log origin/$BASE..HEAD --format='%s'` 로 본다. push 전이면 고치고, push 뒤면 그대로 둔다.
- 아직 push 하지 않았고 `$BASE` 가 앞서 있으면 `git fetch origin && git rebase origin/$BASE` 후 테스트를 다시 돌린다.

## 2. push

```bash
git push -u origin HEAD
```

origin 이 `git@github.com:` 인데 SSH 가 거부되면 원격 설정은 바꾸지 말고 `git push https://github.com/$REPO.git HEAD` 로 올린다.

## 3. PR 열기

제목은 커밋 제목과 같은 형식(커밋이 여럿이면 전체를 아우르는 한 문장), 본문은 한국어로 `## 문제`(기능이면 `## 내용`)·`## 수정`·`## 확인`.

```bash
cat > "$(git rev-parse --git-dir)/PR_BODY.md" <<'EOF'
## 문제
…

## 수정
- …

## 확인
- …
EOF
gh pr create --base "$BASE" --title "🐞 Fix: …한다" --body-file "$(git rev-parse --git-dir)/PR_BODY.md"
```

본문에 도구 서명(`Generated with …`)을 넣지 않는다.

## 4. 클라우드 리뷰 기다리기

PR 을 열거나 push 하면 체크 세 개가 돈다: `Cursor Bugbot`(버그·인라인 지적), `Cursor Security Agent: Security Reviewer`(보안), `Cursor Approval Agent: Pull Request Router and Approver`(위험도 판정 → 자동 승인 또는 사람 리뷰어 배정). 저장소에 GitHub Actions CI 가 있으면 그것도 기다린다.

```bash
gh pr checks <N>
gh pr view <N> --json state,reviewDecision,mergeStateStatus,reviews --jq '{state,reviewDecision,mergeStateStatus,reviews:[.reviews[]|{author:.author.login,state,body:(.body[0:200])}]}'
gh api "repos/$REPO/pulls/<N>/comments" --jq '.[] | {id, path, line, user: .user.login, body: .body[0:400]}'
```

- 수 분 걸린다. 3~5분 간격으로 확인하고, 30분이 지나도 끝나지 않으면 사용자에게 알린다. 짧은 간격으로 계속 조회하지 않는다.
- `Bugbot couldn't run - usage limit reached` 코멘트가 달리거나 Cursor 체크가 아예 붙지 않는 저장소면 7단계(직접 검토)로 간다.
- Approval Agent 가 사람 리뷰어를 배정하면(위험도 medium 이상 등) 사용자에게 알리고 승인을 기다린다.

## 5. 지적 검토·수정

리뷰 스레드 목록:

```bash
gh api graphql -F owner="${REPO%/*}" -F name="${REPO#*/}" -F n=<N> -f query='
query($owner:String!,$name:String!,$n:Int!){repository(owner:$owner,name:$name){pullRequest(number:$n){
  reviewThreads(first:100){nodes{id isResolved comments(first:1){nodes{databaseId author{login} path line body}}}}}}}'
```

지적마다:

1. 지적한 코드와 호출 경로를 읽고, 가능하면 테스트로 재현한다.
2. **맞으면** 고치고 재현 테스트를 더한다. 지적이 여러 건이면 모두 고친 뒤 바뀐 범위 테스트 → 커밋 직전 전체 검사를 한 번만 돌리고 `jdone-commit` 규칙으로 커밋해 push 한다. 스레드에 고친 커밋과 내용을 한국어로 답한다.
3. **틀리면** 근거(코드 위치·조건·의도된 설계)를 한국어로 답한다.
4. 답한 스레드는 resolve 한다.

```bash
gh api "repos/$REPO/pulls/<N>/comments/<comment-id>/replies" -f body='<커밋 해시> 에서 고쳤습니다: …'
gh api graphql -f id=<thread-id> -f query='mutation($id:ID!){resolveReviewThread(input:{threadId:$id}){thread{isResolved}}}'
```

- Bugbot Autofix(작성자 Cursor Agent)가 PR 브랜치에 커밋을 올렸으면 `git pull` 로 받아 내용을 검토하고, 동작이 바뀌었으면 테스트로 고정한다.
- 새로 push 하면 리뷰가 다시 돈다. 4단계로 돌아가 다시 기다린다.

## 6. 머지

체크가 모두 통과하고, 승인(Approval Agent 자동 승인 또는 사람 승인)됐고, 답하지 않은 지적 스레드가 없을 때:

```bash
gh pr view <N> --json state,mergedAt   # 이미 머지됐는지 먼저 확인
gh pr merge <N> --squash
```

- 머지 커밋은 GitHub 기본값(제목 `PR 제목 (#N)`, 본문 각 커밋 메시지)을 그대로 둔다.
- 브랜치가 다른 워크트리에 체크아웃돼 있으면 `--delete-branch` 를 붙이지 않는다 (그 워크트리까지 지운다).

## 7. Cursor 한도일 때 직접 검토

Bugbot 이 사용량 한도로 돌지 않았으면 에이전트가 리뷰를 대신한다.

1. `gh pr diff <N>` 을 바뀐 영역별로 읽고 버그(경계값·동시성·자원 해제·오류 처리·하위 호환)·보안·`AGENTS.md` 스타일 위반을 찾는다. 저장소의 `.cursor/BUGBOT.md` 「특히 볼 것」을 체크리스트로 쓴다.
2. 찾은 것마다 스스로 반박해 본다(정말 그 입력이 들어오는가, 다른 곳에서 막는가). 남은 것만 고친다.
3. PR 에 `직접 검토 결과` 코멘트를 남긴다: 본 범위, 고친 것(커밋), 문제없다고 판단한 것과 이유.
4. 6단계로 머지한다.

## 8. 정리

```bash
git switch "$BASE" && git pull --ff-only
gh pr view <N> --json headRefOid -q .headRefOid   # 브랜치 끝 커밋과 같은지 확인
git branch -D <브랜치>
git push origin --delete <브랜치>    # 원격 브랜치가 남아 있으면
```

- squash 머지라 `git branch -d` 는 머지 안 됨으로 나온다. `gh pr view <N> --json headRefOid` 의 커밋이 브랜치 끝과 같으면 지워도 된다.
- 머지 뒤 몇 분 안에 Bugbot 지적이 늦게 달리기도 한다. 한 번 더 확인하고, 있으면 새 브랜치에서 후속 PR 로 고친다.
- 끝나면 PR 번호·머지 커밋·남은 일(배포 등)을 한국어로 보고한다. 배포는 사용자가 시킬 때만 한다.
