
-- phttp
--- @class phttp
--- @field get fun(url: string, options: table): table
--- @field post fun(url: string, options: table): table
--- @field put fun(url: string, options: table): table
--- @field delete fun(url: string, options: table): table
--- @field patch fun(url: string, options: table): table
--- @field request fun(url: string, options: table): table
--- @field request_batch fun(urls: table, options: table): table

-- json
--- @class json
--- @field encode fun(data: table): string
--- @field decode fun(data: string): table




-- Potato 
--- @class Potato
--- @field kv potato.kv
--- @field cap potato.cap
--- @field db potato.db
--- @field core potato.core

-- KV Module
--- @class potato.kv
--- @field query fun(query: table): table
--- @field add fun(data: table): table
--- @field get fun(group: string, key: string): table
--- @field get_by_group fun(group: string, offset: number, limit: number): table
--- @field remove fun(group: string, key: string): nil
--- @field update fun(group: string, key: string, data: table): nil
--- @field upsert fun(group: string, key: string, data: table): nil

-- Cap Module
--- @class potato.cap
--- @field list fun(): table
--- @field execute fun(action: string, params: table): table
--- @field list_methods fun(): table
--- @field sign_token fun(options: table): string

-- DB Module
--- @class potato.db
--- @field run_ddl fun(ddl: string): nil
--- @field run_query fun(query: string, ...): table
--- @field run_query_one fun(query: string, ...): table
--- @field insert fun(table: string, data: table): number
--- @field update_by_id fun(table: string, id: number, data: table): nil
--- @field delete_by_id fun(table: string, id: number): nil
--- @field find_by_id fun(table: string, id: number): table
--- @field update_by_cond fun(table: string, cond: table, data: table): nil
--- @field delete_by_cond fun(table: string, cond: table): nil
--- @field find_all_by_cond fun(table: string, cond: table): table
--- @field find_one_by_cond fun(table: string, cond: table): table
--- @field find_all_by_query fun(query: table): table
--- @field find_by_join fun(query: table): table
--- @field list_tables fun(): table
--- @field list_columns fun(table: string): table
--- @field start_txn fun(): potato.txn

-- Txn Module
--- @class potato.txn
--- @field run_ddl fun(ddl: string): nil
--- @field run_query fun(query: string, ...): table
--- @field run_query_one fun(query: string, ...): table
--- @field insert fun(table: string, data: table): number
--- @field update_by_id fun(table: string, id: number, data: table): nil
--- @field delete_by_id fun(table: string, id: number): nil
--- @field find_by_id fun(table: string, id: number): table
--- @field update_by_cond fun(table: string, cond: table, data: table): nil
--- @field delete_by_cond fun(table: string, cond: table): nil
--- @field find_all_by_cond fun(table: string, cond: table): table
--- @field find_one_by_cond fun(table: string, cond: table): table
--- @field find_all_by_query fun(query: table): table
--- @field find_by_join fun(query: table): table
--- @field commit fun(): nil
--- @field rollback fun(): nil

-- Core Module
--- @class potato.core
--- @field publish_event fun(options: table): nil
--- @field file_token fun(options: table): string, nil
--- @field sign_advisery_token fun(options: table): string, nil
--- @field parse_advisery_token fun(token: string): table, nil
--- @field read_package_file fun(path: string): string, nil

-- Http Context
--- @class HttpContext
--- @field request fun(): table
--- @field param fun(key: string): string
--- @field type fun(): string -- http
--- @field get_user_claim fun(): table
--- @field get_header fun(key: string): string
--- @field get_json fun(): table
--- @field set_json fun(code: number, target: table): nil

-- Action Context
--- @class ActionContext
--- @field param fun(key: string): string
--- @field type fun(): string -- action
--- @field get_inner_payload fun(): table
--- @field get_inner_value fun(path: string): any
--- @field execute fun(action: string, params: table): table
--- @field list_actions fun(): table