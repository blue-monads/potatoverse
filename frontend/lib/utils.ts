

const KEY = "_potato_login_info_";


export const saveLoginData = (accessToken: string, userInfo: any) => {
     localStorage.setItem(KEY, JSON.stringify({ accessToken, userInfo }));
}

export const getLoginData = () => {
    const item = localStorage.getItem(KEY);
    if (!item) return null;
    return JSON.parse(item);
}

export const removeLoginData = () => {
    localStorage.removeItem(KEY);
}

/**
 * Checks if the current host matches any wildcard host configuration (*.domain).
 * When running under a wildcard domain, each space gets a dedicated subdomain (zz-<spaceId>-<serverKey>.<domain>),
 * preventing namespace collisions between duplicate installations.
 */
export const isWildcardHost = (hosts: string[], currHost: string): boolean => {
    if (!hosts || hosts.length === 0 || !currHost) return false;
    const cleanCurrHost = currHost.split(':')[0].toLowerCase();
    const currHostLower = currHost.toLowerCase();

    for (const host of hosts) {
        const h = host.trim().toLowerCase();
        if (h.includes('*')) {
            const base = h.replace('*.', '');
            if (
                currHostLower.startsWith(base) ||
                cleanCurrHost === base ||
                cleanCurrHost.endsWith('.' + base)
            ) {
                return true;
            }
        }
    }
    return false;
};


