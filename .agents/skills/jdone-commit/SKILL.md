---
name: jdone-commit
description: JDONE 커밋 규칙(「<이모지> <타입>: 한국어 한 문장」 제목 + [영역] 본문)으로 변경을 검증하고 커밋 메시지를 써서 커밋한다. 커밋, 커밋 메시지 작성, 변경 정리를 요청받았을 때 쓴다.
---

# JDONE 커밋

규칙의 원문은 저장소 루트 `AGENTS.md` 의 「커밋 메시지」다. 이 스킬은 그 규칙대로 커밋하는 절차다.

## 1. 상태 확인

```bash
git branch --show-current
git status --short
git diff --stat && git diff --staged --stat
```

- 기본 브랜치(`main`, gortsplib 는 `fix/digest-qop-auth`)에 있으면 커밋하지 말고 `<타입>/<짧은-영어-설명>` 브랜치를 만든다. 사용자가 "main 에 바로" 라고 한 경우만 예외.
- 이번 작업과 관계없는 변경, 비밀 값(`.env`·키·비밀번호), 산출물(`dist/`·캐시·로그·모델 파일)이 섞였으면 커밋에서 빼고 사용자에게 알린다.
- 목적이 둘 이상이면 커밋을 나눈다. 한 파일에 두 목적이 섞여 나누기 어려우면 사용자에게 묻는다.

## 2. 검토 한 번 → 전체 검사 한 번

작업 중에는 `AGENTS.md` 「명령」의 **바뀐 범위 테스트**만 돌렸다는 전제다. 커밋 직전에 아래 두 가지를 한 번씩만 한다.

1. **검토**: `git diff $(git merge-base origin/HEAD HEAD)` 를 처음부터 한 번 읽고 「코드 스타일」과 `.cursor/BUGBOT.md` 「특히 볼 것」 점검표로 확인해 고친다. 고친 곳은 바뀐 범위 테스트로 확인한다.
2. **전체 검사**: 「명령」의 「커밋 직전 전체 검사」를 돌린다. 오래 걸리면 백그라운드로 돌리고 그동안 3단계(메시지)를 쓴다.
   - 실패하면 고친 뒤 실패한 것만 다시 돌린다(`pytest --lf`, `go test -run`). 고친 곳이 여러 모듈이 쓰는 공용 코드일 때만 전체를 다시 돌린다.
   - 실패가 남으면 커밋하지 말고 결과를 보고한다.
   - 문서·주석만 바꿨으면 둘 다 생략한다.

## 3. 메시지 쓰기

1. **타입**: `AGENTS.md` 표의 21가지 중 하나 — `✨ Feat` `🐞 Fix` `🧹 Chore` `♻️ Refactor` `📝 Docs` `🧪 Test` `🚀 Perf` `⚙️ Config` `🚑️ Hotfix` `📦 Deploy` `📁 Assets` `📌 Package` `🗄️ DB` `🔥 Remove` `🔀 Merge` `⏪ Revert` `🔒 Security` `📊 Analytics` `🎨 Style` `🚧 WIP` `🎉 Init`. 목록에 없는 이모지는 쓰지 않는다.
2. **제목**: 무엇이 어떻게 달라지는지 한국어 한 문장, "~한다"로 끝낸다. 코드 이름보다 동작·결과. 마침표와 `(#N)` 은 쓰지 않는다.
3. **빈 줄** 한 줄 (없으면 git 이 본문까지 제목으로 읽는다).
4. **배경 단락**(선택): 요구사항 ID, 현상, 원인을 평서문으로.
5. **[영역] 섹션**: `AGENTS.md` 에 적힌 이 저장소의 영역 이름을 쓴다. diff 의 파일마다
   - `- 파일 이름: 바꾼 내용` (경로 없이 파일 이름, 새 파일은 `(신규)`, 개조식 …함·…음·명사 끝)
   - 다음 줄에 두 칸 들여 `(이유·근거·수치·버린 대안)`
   - 테스트는 `[Test]`, 문서는 맨 끝 `[Docs]`.
6. 이유는 diff·대화·실측으로 확인한 것만 쓴다. 모르는 이유를 지어내지 않는다.
7. `Co-Authored-By:`·`Generated with …` 같은 도구 서명 줄은 넣지 않는다.

한두 줄짜리 작은 변경은 제목만 써도 된다.

## 4. 커밋

줄바꿈과 괄호가 그대로 남도록 파일로 넘긴다.

```bash
git add <파일...>
MSG="$(git rev-parse --git-dir)/COMMIT_DRAFT"
cat > "$MSG" <<'EOF'
🐞 Fix: …한다

[Backend]
- 파일.py: …함
  (…)

[Test]
- test_파일.py: …
EOF
git commit -F "$MSG"
git log -1 --format=%B
```

- push 한 커밋은 amend·rebase·force push 하지 않는다. 고칠 것은 새 커밋으로 더한다.
- 사용자가 PR 까지 맡겼으면 `jdone-pr` 스킬로 이어 간다.
