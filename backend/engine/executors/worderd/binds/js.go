package binds

const PotatoJs = `
const call = async (requestId, module, method, ...params) => {
    const resp = await fetch('http://internal_bindings/call', {
        method: 'POST',
        headers: {
            'X-Potato-Request-ID': requestId || ''
        },
        body: JSON.stringify({ module, method, params })
    });
    const data = await resp.json();
    if (data.error) {
        throw new Error(data.error);
    }
    return data.result;
};

export const usePotato = (request) => {
    const requestId = request.headers.get('X-Potato-Request-ID');
    return {
        db: {
            run_query: (query, ...args) => call(requestId, 'db', 'run_query', query, ...args),
            insert: (table, data) => call(requestId, 'db', 'insert', table, data),
            update_by_id: (table, id, data) => call(requestId, 'db', 'update_by_id', table, id, data),
            delete_by_id: (table, id) => call(requestId, 'db', 'delete_by_id', table, id),
            find_by_id: (table, id) => call(requestId, 'db', 'find_by_id', table, id),
            find_all_by_cond: (table, cond) => call(requestId, 'db', 'find_all_by_cond', table, cond),
            find_one_by_cond: (table, cond) => call(requestId, 'db', 'find_one_by_cond', table, cond),
        },
        kv: {
            get: (group, key) => call(requestId, 'kv', 'get', group, key),
            set: (group, key, value) => call(requestId, 'kv', 'set', group, key, value),
            remove: (group, key) => call(requestId, 'kv', 'remove', group, key),
        },
        signer: {
            parse_space: (token) => call(requestId, 'signer', 'parse_space', token),
        }
    };
};
`
