#!/usr/bin/env python3
import argparse
import json
import sys
from typing import Any


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="RBAC matrix helper")
    sub = p.add_subparsers(dest="cmd", required=True)

    v = sub.add_parser("validate-target-users", help="Normalize TARGET_USERS_JSON")
    v.add_argument("--raw", required=True, help="Raw TARGET_USERS_JSON string")

    m = sub.add_parser("build-matrix", help="Build user/menu matrix JSON")
    m.add_argument("--users-json", required=True, help="admin_users API response JSON")
    m.add_argument("--menus-json", required=True, help="rbac menus API response JSON")
    m.add_argument("--targets-json", required=True, help="normalized target users JSON")
    m.add_argument("--only-menu-keys", default="", help="CSV menu_key filter")

    c = sub.add_parser("to-cases-tsv", help="Convert matrix JSON to TSV")
    c.add_argument("--matrix-json", required=True, help="Matrix JSON")

    return p.parse_args()


def try_parse_json(s: str) -> tuple[Any | None, Exception | None]:
    try:
        return json.loads(s), None
    except Exception as e:  # noqa: BLE001
        return None, e


def validate_target_users(raw: str) -> int:
    fallback = [{"username": "admin", "password": "admin123"}]
    s = (raw or "").strip()
    candidates = []
    if s:
        candidates.append(s)
        if len(s) >= 2 and ((s[0] == "'" and s[-1] == "'") or (s[0] == '"' and s[-1] == '"')):
            candidates.append(s[1:-1].strip())
        candidates.append(s.replace('\\"', '"'))
        if len(s) >= 2 and ((s[0] == "'" and s[-1] == "'") or (s[0] == '"' and s[-1] == '"')):
            t = s[1:-1].strip()
            candidates.append(t.replace('\\"', '"'))

    obj = None
    last_err: Exception | None = None
    for c in candidates:
        if not c:
            continue
        parsed, err = try_parse_json(c)
        if err is None:
            obj = parsed
            break
        last_err = err

    if obj is None:
        print(f"WARN: invalid TARGET_USERS_JSON, fallback to default. parse_error={last_err}", file=sys.stderr)
        print("      example: TARGET_USERS_JSON='[{\"username\":\"admin\",\"password\":\"admin123\"}]'", file=sys.stderr)
        obj = fallback

    if not isinstance(obj, list):
        print("WARN: TARGET_USERS_JSON root is not array, fallback to default", file=sys.stderr)
        obj = fallback

    for i, it in enumerate(obj):
        if not isinstance(it, dict):
            print(f"WARN: TARGET_USERS_JSON item#{i} is not object, fallback to default", file=sys.stderr)
            obj = fallback
            break
        if "username" not in it or "password" not in it:
            print(f"WARN: TARGET_USERS_JSON item#{i} missing username/password, fallback to default", file=sys.stderr)
            obj = fallback
            break

    print(json.dumps(obj, ensure_ascii=False))
    return 0


def build_matrix(users_json: str, menus_json: str, targets_json: str, only_menu_keys: str) -> int:
    users_obj = json.loads(users_json)
    menus_obj = json.loads(menus_json)
    targets = json.loads(targets_json)
    menu_key_filter = {x.strip() for x in only_menu_keys.split(",") if x.strip()}

    users_map: dict[str, dict[str, Any]] = {}
    for u in users_obj.get("users") or []:
        if isinstance(u, dict) and u.get("username"):
            users_map[u["username"]] = {"id": int(u.get("id") or 0), "username": u["username"]}

    out_targets = []
    for t in targets:
        if not isinstance(t, dict):
            continue
        un = str(t.get("username") or "").strip()
        pw = str(t.get("password") or "").strip()
        if not un or not pw:
            continue
        if un not in users_map or users_map[un]["id"] <= 0:
            continue
        out_targets.append({"username": un, "password": pw, "id": users_map[un]["id"]})

    out_menus = []
    for m in menus_obj.get("menus") or []:
        if not isinstance(m, dict):
            continue
        mk = str(m.get("menu_key") or "").strip()
        path = str(m.get("path") or "").strip()
        kind = int(m.get("kind") or 0)
        if not mk or not path or kind != 1:
            continue
        if menu_key_filter and mk not in menu_key_filter:
            continue
        out_menus.append(
            {
                "id": int(m.get("id") or 0),
                "menu_key": mk,
                "path": path,
                "label": str(m.get("label") or "").strip(),
            }
        )

    print(json.dumps({"targets": out_targets, "menus": out_menus}, ensure_ascii=False))
    return 0


def to_cases_tsv(matrix_json: str) -> int:
    m = json.loads(matrix_json)
    for t in m.get("targets") or []:
        for menu in m.get("menus") or []:
            print(
                "\t".join(
                    [
                        str(t["id"]),
                        t["username"],
                        t["password"],
                        menu["menu_key"],
                        menu["path"],
                        str(menu["id"]),
                    ]
                )
            )
    return 0


def main() -> int:
    args = parse_args()
    if args.cmd == "validate-target-users":
        return validate_target_users(args.raw)
    if args.cmd == "build-matrix":
        return build_matrix(args.users_json, args.menus_json, args.targets_json, args.only_menu_keys)
    if args.cmd == "to-cases-tsv":
        return to_cases_tsv(args.matrix_json)
    return 1


if __name__ == "__main__":
    raise SystemExit(main())

