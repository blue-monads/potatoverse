import { deriveHost } from '@/lib/api';


const buildIframeSrc = (nskey: string, host: string) => {

    let src = `/zz/space/${nskey}`;

    if (host) {
        // Clean the host - remove any protocol, paths, or leading/trailing slashes
        let cleanHost = host.trim();

        // Remove protocol if present
        cleanHost = cleanHost.replace(/^https?:\/\//, '');

        // Remove any path that might be included (take only the hostname part)
        // This prevents issues like "hostname/path" becoming part of the URL
        cleanHost = cleanHost.split('/')[0];

        // Remove trailing slashes
        cleanHost = cleanHost.replace(/\/+$/, '');

        // Determine protocol and port from current origin
        const origin = window.location.origin;
        const isSecure = origin.startsWith("https://");

        // Extract port from current origin (needed for localhost development)
        let port = '';
        try {
            const url = new URL(origin);
            const originPort = url.port;
            // Only include port if it's non-standard (not 80 for http, not 443 for https)
            if (originPort && originPort !== '' &&
                ((!isSecure && originPort !== '80') || (isSecure && originPort !== '443'))) {
                port = `:${originPort}`;
            }
        } catch (e) {
            // Fallback: try to extract port manually if URL parsing fails
            const portMatch = origin.match(/:(\d+)$/);
            if (portMatch) {
                const originPort = portMatch[1];
                if ((!isSecure && originPort !== '80') || (isSecure && originPort !== '443')) {
                    port = `:${originPort}`;
                }
            }
        }

        // Build the URL - preserve port from current origin for localhost development
        src = `${isSecure ? "https://" : "http://"}${cleanHost}${port}/zz/space/${nskey}`;
    }

    return src;
}


export const deriveHostAndIframeSrc = async (namespace: string, spaceId: string) => {
    const resp = await deriveHost(namespace, spaceId);
    if (resp.status !== 200) {
        console.error("failed to derive host");
        return;
    }

    console.log("@deriveHostAndIframeSrc", namespace, spaceId, resp.data)


    return buildIframeSrc(namespace, resp.data.host)

}

