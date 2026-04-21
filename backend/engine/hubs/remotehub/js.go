package remotehub

const PotatoJs = `
const request = async (service, token, path, method, body) => {
    const url = 'http://127.0.0.1/zz/rt_binds/' + path;
    const options = {
        method,
        headers: {
            'X-Exec-Header': token,
            'Content-Type': 'application/json'
        },
        body: body ? JSON.stringify(body) : undefined
    };

    const resp = service ? await service.fetch(url, options) : await fetch(url, options);
    
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

const get = (service, token, path) => request(service, token, path, 'GET');
const post = (service, token, path, body) => request(service, token, path, 'POST', body);

export const usePotato = (req, env) => {
    const token = req.headers.get('X-Exec-Header');
    const service = env ? env.internal_bindings : null;
    return {
        db: {
            run_query: (query, ...args) => post(service, token, 'db/run_query', { query, args }),
            insert: (table, data) => post(service, token, 'db/insert', { table, data }),
            update_by_id: (table, id, data) => post(service, token, 'db/update_by_id', { table, id, data }),
            delete_by_id: (table, id) => post(service, token, 'db/delete_by_id', { table, id }),
            find_by_id: (table, id) => post(service, token, 'db/find_by_id', { table, id }),
            find_all_by_cond: (table, cond) => post(service, token, 'db/find_all_by_cond', { table, cond }),
            find_one_by_cond: (table, cond) => post(service, token, 'db/find_one_by_cond', { table, cond }),
        },
        kv: {
            get: (group, key) => get(service, token, 'kv/' + group + '/' + key),
            set: (group, key, value) => post(service, token, 'kv/upsert', { group, key, data: value }),
            remove: (group, key) => post(service, token, 'kv/remove', { group, key }),
        },
        core: {
            publish_event: (name, payload, resource_id, collapse_key) => 
                post(service, token, 'core/publish_event', { name, payload, resource_id, collapse_key }),
            get_env: (key) => get(service, token, 'core/env/' + key),
            read_package_file: (path) => get(service, token, 'core/read_package_file/' + path),
            list_files: (path) => get(service, token, 'core/list_files/' + path),
        }
    };
};
`
