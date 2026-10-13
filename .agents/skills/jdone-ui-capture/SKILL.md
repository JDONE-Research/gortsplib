---
name: jdone-ui-capture
description: 화면(UI)이 바뀐 작업에서 바뀐 화면을 Playwright(없으면 Claude Code·Codex 내장 브라우저)로 1920×1080 캡처해 ~/Downloads/JDONE-yyMMdd-작업명/01-기능/01-세부.png 로 정리한다. 프론트 화면·화면 문구가 바뀐 기능을 확인할 때, 사용자가 화면 캡처를 요청할 때 쓴다. 백엔드만 바뀌고 화면이 그대로인 작업에는 쓰지 않는다.
---

# JDONE 화면 캡처

규칙의 원문은 저장소 루트 `AGENTS.md` 의 「화면 캡처」다.

## 1. 할지 정하기

- **한다**: 프론트 코드·화면 문구가 바뀜, 백엔드 변경이 화면에 보이는 값·목록·오류 문구를 바꿈, 사용자가 캡처를 요청함.
- **안 한다**: 백엔드만 바뀌고 화면이 그대로, 테스트·문서·설정만 바뀜.

## 2. 폴더와 이름

```text
~/Downloads/JDONE-yyMMdd-작업명/
  01-기능1/
    01-기능세부명.png
    02-기능세부명.png
  02-기능2/
    01-기능세부명.png
```

- `yyMMdd` 는 캡처한 날(예: `261014`), 작업명은 짧은 한국어로 띄어쓰기 대신 하이픈을 쓴다 (예: `JDONE-261014-카메라-일괄-등록`).
- 번호는 두 자리. 기능 폴더는 화면 메뉴 순서, 파일은 화면 흐름 순서(빈 화면 → 입력 → 결과 → 오류)로 매긴다.
- 파일 이름에 폴더와 겹치는 기능 이름이나 단계 접두어를 넣지 않는다. 라이트·다크가 달라 보이는 화면만 `-다크` 를 붙인 장을 라이트 바로 다음 번호로 더한다.
- 같은 작업을 이어서 찍으면 그 폴더의 마지막 번호 다음부터 쓴다. 사용자가 지운 폴더·파일은 확인이 끝난 것이니 다시 만들지 않는다.

## 3. 화면 띄우기

- `AGENTS.md` 「명령」대로 dev 서버를 띄운다(예: `pnpm dev`). 백엔드는 `VITE_DEV_API_TARGET` 의 시험 서버나 로컬 가짜 API 를 쓴다.
- 로그인이 필요하면 사용자에게 로그인을 부탁하거나 프로젝트의 시험 계정을 쓴다. 비밀번호·토큰을 파일·로그·캡처에 남기지 않는다.
- 가짜 데이터는 실제와 어긋나지 않게(건수·상태·시각이 서로 맞게) 넣는다.
- 화면에 버그·잘못된 탐지·현장 실데이터(해군 영상 등)·내부 주소·비밀값이 보이면 찍지 않는다. 고친 뒤 다시 찍거나 사용자에게 알린다.

## 4. 찍기

뷰포트는 1920×1080, 배율은 CSS 픽셀(1배)이다. 기능을 확인할 때마다 바로 저장하고, 마지막에 몰아서 찍지 않는다.

1. **Playwright (먼저)**
   - **Playwright MCP 도구**가 있으면 `browser_resize`(1920, 1080) → 화면 이동 → `browser_take_screenshot`(`scale: "css"`, `filename: ".playwright-mcp/JDONE-yyMMdd-작업명/01-기능/01-세부명.png"`). MCP 는 작업 폴더 안에만 저장하므로, 저장한 뒤 다운로드로 옮기고 작업 폴더에서는 지운다.

     ```bash
     mkdir -p ~/Downloads && rsync -a --remove-source-files .playwright-mcp/JDONE-yyMMdd-작업명 ~/Downloads/ && rm -rf .playwright-mcp/JDONE-yyMMdd-작업명
     ```

   - **스크립트**로 찍을 때는 저장소 밖 임시 폴더에서 한다.

     ```bash
     cd "$(mktemp -d)" && npm i --silent playwright && npx playwright install chromium   # 브라우저가 이미 있으면 건너뜀
     node capture.mjs
     ```

     ```js
     // capture.mjs
     import { mkdirSync } from "node:fs";
     import { homedir } from "node:os";
     import { join } from "node:path";

     import { chromium } from "playwright";

     const root = join(homedir(), "Downloads", "JDONE-261014-카메라-일괄-등록");

     const shot = async (page, feature, name) => {
       const dir = join(root, feature);
       mkdirSync(dir, { recursive: true });
       await page.screenshot({ path: join(dir, `${name}.png`) });
     };

     const browser = await chromium.launch();
     const page = await browser.newPage({
       viewport: { width: 1920, height: 1080 },
     });
     await page.goto("http://localhost:18000/cameras");
     await shot(page, "01-일괄-등록", "01-빈-목록");
     await browser.close();
     ```

2. **내장 브라우저 (Playwright 가 없을 때)**: Claude Code 내장 브라우저(Claude Browser)나 Codex 내장 브라우저에서 뷰포트를 1920×1080 으로 맞추고 같은 순서로 찍는다. 그 도구가 이미지를 파일로 저장하지 못하면 화면 확인만 하고, 저장은 위 스크립트 방식으로 한다.
3. 둘 다 안 되면 찍지 못한 화면 목록을 사용자에게 알린다.

## 5. 정리

- 임시 스크립트·`.playwright-mcp/` 의 캡처는 지운다. 캡처·스크립트를 저장소에 커밋하지 않는다.
- 끝나면 폴더 경로와 기능별 캡처 목록을 보고하고, PR 「확인」에 캡처 폴더 이름을 적는다.
