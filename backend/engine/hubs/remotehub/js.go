package remotehub

const PotatoJs = `
const BASE_URL = 'http://internal_bindings/zz/rt_binds';

const request = async (token, path, method, body) => {
    const resp = await fetch(BASE_URL + '/' + path, {
        method,
        headers: {
            'X-Exec-Header': token,
            'Content-Type': 'application/json'
        },
        body: body ? JSON.stringify(body) : undefined
    });
    
    if (!resp.ok) {
        let errorText = await resp.text();
        try {
            const errorJson = JSON.parse(errorText);
            throw new Error(errorJson.error || errorJson.message || 'HTTP error! status: ' + resp.status);
        } catch (e) {
            if (e instanceof SyntaxError) {
                throw new Error('HTTP error! status: ' + resp.status + ', message: ' + errorText);
            }
            throw e;
        }
    }

    const data = await resp.json();
    if (data.error) {
        throw new Error(data.error);
    }
    return data.result !== undefined ? data.result : data;
};

const get = (token, path) => request(token, path, 'GET');
const post = (token, path, body) => request(token, path, 'POST', body);

export const usePotato = (req) => {
    const token = req.headers.get('X-Exec-Header');
    return {
        db: {
            run_query: (query, ...args) => post(token, 'db/run_query', { query, args }),
            insert: (table, data) => post(token, 'db/insert', { table, data }),
            update_by_id: (table, id, data) => post(token, 'db/update_by_id', { table, id, data }),
            delete_by_id: (table, id) => post(token, 'db/delete_by_id', { table, id }),
            find_by_id: (table, id) => post(token, 'db/find_by_id', { table, id }),
            find_all_by_cond: (table, cond) => post(token, 'db/find_all_by_cond', { table, cond }),
            find_one_by_cond: (table, cond) => post(token, 'db/find_one_by_cond', { table, cond }),
        },
        kv: {
            get: (group, key) => get(token, 'kv/' + group + '/' + key),
            set: (group, key, value) => post(token, 'kv/upsert', { group, key, data: value }),
            remove: (group, key) => post(token, 'kv/remove', { group, key }),
        },
        core: {
            publish_event: (name, payload, resource_id, collapse_key) => 
                post(token, 'core/publish_event', { name, payload, resource_id, collapse_key }),
            get_env: (key) => get(token, 'core/env/' + key),
            read_package_file: (path) => get(token, 'core/read_package_file/' + path),
            list_files: (path) => get(token, 'core/list_files/' + path),
        }
    };
};
`
