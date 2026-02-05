#!/bin/bash

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
OUT="$SCRIPT_DIR/../.custom_gpt"
mkdir -p "$OUT"

WEAVER_VERSION="v0.1.4"
go install "github.com/aatuh/weaver/cmd/weaver@${WEAVER_VERSION}"

WEAVER_BIN="weaver"

"$WEAVER_BIN" -include-tree-compact -skip-binary \
    -blacklist .gitignore -root "$SCRIPT_DIR/../docs/" > "$OUT/docs_contents.txt"
"$WEAVER_BIN" -include-tree-compact -skip-binary \
    -blacklist .gitignore -root "$SCRIPT_DIR/../api/" > "$OUT/recsys-svc_contents.txt"
"$WEAVER_BIN" -include-tree-compact -skip-binary \
    -blacklist .gitignore -root "$SCRIPT_DIR/../recsys-algo/" > "$OUT/recsys-algo_contents.txt"
"$WEAVER_BIN" -include-tree-compact -skip-binary \
    -blacklist .gitignore -root "$SCRIPT_DIR/../recsys-eval/" > "$OUT/recsys-eval_contents.txt"
"$WEAVER_BIN" -include-tree-compact -skip-binary \
    -blacklist .gitignore -root "$SCRIPT_DIR/../recsys-pipelines/" > "$OUT/recsys-pipelines_contents.txt"
"$WEAVER_BIN" -include-tree-compact -skip-binary \
    -blacklist scripts/custom_gpt_extract_blacklist -blacklist .gitignore -root "$SCRIPT_DIR/../" > "$OUT/root_contents.txt"
