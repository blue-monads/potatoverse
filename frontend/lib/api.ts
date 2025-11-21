import axios, { AxiosInstance } from "axios";
import { getLoginData } from "./utils";


let iaxios: AxiosInstance = axios.create({
    baseURL: "/zz/api",
});

export const initHttpClient = () => {

    const data = getLoginData();

    

    const headers: Record<string, string> = {
        "Content-Type": "application/json",
        "X-Clacks-Overhead": "Aaron Swartz",
    }

    if (data?.accessToken) {
        headers["Authorization"] = `TokenV1 ${data.accessToken}`;
    }


    iaxios = axios.create({
        baseURL: "/zz/api",
        headers,
    });


}


export const login = async (username: string, password: string) => {
    return iaxios.post<{ access_token: string, user_info: User }>("/core/auth/login", {
        username,
        password,
    });
}

export interface User {
    id: number;
    name: string;
    utype: string;
    ugroup: string;
    email: string;
    phone: string;
    username: string;
    bio: string;
    password: string;
    is_verified: boolean;
    extrameta: string;
    createdAt: string;
    owner_user_id: number;
    owner_space_id: number;
    msg_read_head: number;
    disabled: boolean;
    is_deleted: boolean;
}


// /zz/api/core/user/

export const getUsers = async () => {
    return iaxios.get<User[]>("/core/user");
}

export const getUser = async (id: number) => {
    return iaxios.get<User>(`/core/user/${id}`);
}

// User Invites
export interface UserInvite {
    id: number;
    email: string;
    role: string;
    status: string;
    invited_by: number;
    invited_as_type: string;
    expires_on: string;
    created_at: string;
}

export interface UserInviteResponse extends UserInvite {
    invite_url?: string;
}

export const getUserInvites = async () => {
    return iaxios.get<UserInvite[]>("/core/user/invites");
}

export const getUserInvite = async (id: number) => {
    return iaxios.get<UserInvite>(`/core/user/invites/${id}`);
}

export const createUserInvite = async (data: {
    email: string;
    role: string;
    invited_as_type: string;
}) => {
    return iaxios.post<UserInviteResponse>("/core/user/invites", data);
}

export const updateUserInvite = async (id: number, data: any) => {
    return iaxios.put(`/core/user/invites/${id}`, data);
}

export const deleteUserInvite = async (id: number) => {
    return iaxios.delete(`/core/user/invites/${id}`);
}

export const resendUserInvite = async (id: number) => {
    return iaxios.post<UserInviteResponse>(`/core/user/invites/${id}/resend`);
}

// Invite Acceptance
export interface InviteInfo {
    email: string;
    role: string;
    expires_on: string;
}

export const getInviteInfo = async (token: string) => {
    return iaxios.get<InviteInfo>(`/core/auth/invite/${token}`);
}

export const acceptInvite = async (token: string, data: {
    name: string;
    username: string;
    password: string;
}) => {
    return iaxios.post<{ message: string; user: User }>(`/core/auth/invite/${token}`, data);
}

// Create User Directly
export const createUserDirectly = async (data: {
    name: string;
    email: string;
    username: string;
    utype: string;
    ugroup: string;
}) => {
    return iaxios.post<User>("/core/user/create", data);
}

// User Groups API
export interface UserGroup {
    name: string;
    info: string;
    created_at?: string;
    updated_at?: string;
}

export const getUserGroups = async () => {
    return iaxios.get<UserGroup[]>("/core/user/groups");
}

export const getUserGroup = async (name: string) => {
    return iaxios.get<UserGroup>(`/core/user/groups/${name}`);
}

export const createUserGroup = async (data: {
    name: string;
    info: string;
}) => {
    return iaxios.post<UserGroup>("/core/user/groups", data);
}

export const updateUserGroup = async (name: string, data: {
    info: string;
}) => {
    return iaxios.put<UserGroup>(`/core/user/groups/${name}`, data);
}

export const deleteUserGroup = async (name: string) => {
    return iaxios.delete<void>(`/core/user/groups/${name}`);
}


export interface AdminPortalData {
    popular_keywords: string[] 
    favorite_projects: any[]    
}


export const getAdminPortalData = async (portal_type: string) => {
    return iaxios.get<AdminPortalData>(`/core/self/portalData/${portal_type}`);
}

// Self API
export const getSelfInfo = async () => {
    return iaxios.get<User>("/core/self/info");
}

export const updateSelfBio = async (bio: string) => {
    return iaxios.put<{ message: string }>("/core/self/bio", { bio });
}


export const installPackage = async (url: string) => {
    return iaxios.post<{ package_id: number }>(`/core/package/install`, { url });
}

export const installPackageZip = async (zip: ArrayBuffer) => {
    return iaxios.post<{ package_id: number }>(`/core/package/install/zip`, zip, {
        headers: {
            "Content-Type": "application/zip",
        },
    });
}

export const installPackageEmbed = async (name: string, repoSlug?: string) => {
    return iaxios.post<{ package_id: number }>(`/core/package/install/embed`, { 
        name,
        repo_slug: repoSlug
    });
}

export const deletePackage = async (id: number) => {
    return iaxios.delete<void>(`/core/package/${id}`);
}

export const generatePackageDevToken = async (packageId: number) => {
    return iaxios.post<{ token: string }>(`/core/package/${packageId}/dev-token`);
}

export interface EPackage {
    name: string;
    info: string;
    type: string;
    tags: string;
    format_version: string;
    author_name: string;
    author_email: string;
    author_site: string;
    source_code: string;
    license: string;
    update_url: string;
    artifacts: any[];
    version: string;

}

export const listEPackages = async (repoSlug?: string) => {
    const params = repoSlug ? `?repo=${encodeURIComponent(repoSlug)}` : '';
    return iaxios.get<EPackage[]>(`/core/package/list${params}`);
}

export interface Repo {
    url: string;
    type: string; // http, embeded
    slug: string;
    name: string;
}

export const listRepos = async () => {
    return iaxios.get<Repo[]>(`/core/repo/list`);
}


export interface Package {
    id: number;
    install_id: number;
    name: string;
    description: string;
    info: string;
    slug: string;
    type: string;
    tags: string;
    version: string;
}

export interface Space {
    id: number;
    name: string;
    namespace_key: string;
    owns_namespace: boolean;
    install_id: number;
    executor_type: string;
    sub_type: string;
    owned_by: number;
    extrameta: string;
    is_initilized: boolean;
    is_public: boolean;
}

export interface InstalledSpace {
    spaces: Space[]
    packages: Package[]
}

export const listInstalledSpaces = async () => {
    return iaxios.get<InstalledSpace>(`/core/space/installed`);
}

export interface InstalledPackageInfo {
    installed_package: {
        id: number;
        name: string;
        install_repo: string;
        update_url: string;
        storage_type: string;
        active_install_id: number;
        installed_by: number;
        installed_at?: string;
    };
    spaces: Space[];
    package_versions: PackageVersion[];
}

export interface PackageVersion {
    id: number;
    install_id: number;
    name: string;
    slug: string;
    info: string;
    tags: string;
    format_version: string;
    author_name: string;
    author_email: string;
    author_site: string;
    source_code: string;
    license: string;
    version: string;
}

export const getInstalledPackageInfo = async (packageId: number) => {
    return iaxios.get<InstalledPackageInfo>(`/core/package/${packageId}/info`);
}


export interface SpaceInfo {
    id: number;
    namespace_key: string;
    owns_namespace: boolean;
    package_name: string;
    package_info: string;
}

export const getSpaceInfo = async (space_key: string, host_name: string) => {
    const encodedParams = encodeURIComponent(host_name);
    return iaxios.get<SpaceInfo>(`/core/engine/space_info/${space_key}?host_name=${encodedParams.toString()}`);
}

export interface DeriveHostResponse {
    host: string;
    space_id: number;
}

export const deriveHost = async (nskey: string, spaceId?: string) => {
    const hostname = window.location.hostname;

    return iaxios.get<DeriveHostResponse>(
        `/core/engine/derivehost/${nskey}?host_name=${hostname}&space_id=${spaceId}`
    );
}


export const authorizeSpace = async (space_key: string, space_id: number) => {
    return iaxios.post<{ token: string }>(`/core/space/authorize/${space_key}`, { space_id });
}

// Package Files API
export interface PackageFile {
    id: number;
    package_id: number;
    name: string;
    is_folder: boolean;
    path: string;
    size: number;
    mime: string;
    hash: string;
    store_type: number;
    created_by: number;
    created_at: string;
}



export const listPackageFiles = async (packageId: number, path: string = '' ) => {
    return iaxios.get<PackageFile[]>(`/core/vpackage/${packageId}/files`, {
        params: {
            path,
        },
    });
}

export const getPackageFile = async (packageId: number, fileId: number) => {
    return iaxios.get<PackageFile>(`/core/vpackage/${packageId}/files/${fileId}`);
}

export const downloadPackageFile = async (packageId: number, fileId: number) => {
    return iaxios.get(`/core/vpackage/${packageId}/files/${fileId}/download`, {
        responseType: 'blob'
    });
}

export const deletePackageFile = async (packageId: number, fileId: number) => {
    return iaxios.delete(`/core/vpackage/${packageId}/files/${fileId}`);
}

export const uploadPackageFile = async (packageId: number, file: File, path: string = '') => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);
    
    return iaxios.post(`/core/vpackage/${packageId}/files/upload`, formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
}

export const updatePackageFileContent = async (packageId: number, fileId: number, content: string, fileName: string, path: string = '') => {
    // Delete the old file first
    await deletePackageFile(packageId, fileId);
    
    // Create a blob from the content string
    const blob = new Blob([content], { type: 'text/plain' });
    const file = new File([blob], fileName, { type: 'text/plain' });
    
    // Upload the new file with the same path and name
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);
    
    return iaxios.post(`/core/vpackage/${packageId}/files/upload`, formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
}

// Space KV API
export interface SpaceKV {
    id: number;
    key: string;
    group: string;
    value?: string; // Optional - not included in list responses to reduce payload size
    space_id?: number;
    install_id?: number;
    tag1?: string;
    tag2?: string;
    tag3?: string;
}

export const listSpaceKV = async (installId: number, offset: number = 0, limit: number = 100) => {
    return iaxios.get<SpaceKV[]>(`/core/space/${installId}/kv`, {
        params: {
            offset,
            limit,
        },
    });
}

export const getSpaceKV = async (installId: number, id: number) => {
    return iaxios.get<SpaceKV>(`/core/space/${installId}/kv/${id}`);
}

export const createSpaceKV = async (installId: number, data: Partial<SpaceKV>) => {
    return iaxios.post<SpaceKV>(`/core/space/${installId}/kv`, data);
}

export const updateSpaceKV = async (installId: number, id: number, data: Partial<SpaceKV>) => {
    return iaxios.put<SpaceKV>(`/core/space/${installId}/kv/${id}`, data);
}

export const deleteSpaceKV = async (installId: number, id: number) => {
    return iaxios.delete<void>(`/core/space/${installId}/kv/${id}`);
}

// Space Files API
export interface SpaceFile {
    id: number;
    name: string;
    is_folder: boolean;
    path: string;
    size: number;
    mime: string;
    hash: string;
    storeType: number;
    owner_space_id: number;
    created_by: number;
    created_at: string;
}

export const listSpaceFiles = async (installId: number, path: string = '') => {
    return iaxios.get<SpaceFile[]>(`/core/space/${installId}/files`, {
        params: {
            path,
        },
    });
}

export const getSpaceFile = async (installId: number, fileId: number) => {
    return iaxios.get<SpaceFile>(`/core/space/${installId}/files/${fileId}`);
}

export const downloadSpaceFile = async (installId: number, fileId: number) => {
    return iaxios.get(`/core/space/${installId}/files/${fileId}/download`, {
        responseType: 'blob'
    });
}

export const deleteSpaceFile = async (installId: number, fileId: number) => {
    return iaxios.delete(`/core/space/${installId}/files/${fileId}`);
}

export const uploadSpaceFile = async (installId: number, file: File, path: string = '') => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);
    
    return iaxios.post(`/core/space/${installId}/files/upload`, formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
}

export const updateSpaceFileContent = async (installId: number, fileId: number, content: string, fileName: string, path: string = '') => {
    // Delete the old file first
    await deleteSpaceFile(installId, fileId);
    
    // Create a blob from the content string
    const blob = new Blob([content], { type: 'text/plain' });
    const file = new File([blob], fileName, { type: 'text/plain' });
    
    // Upload the new file with the same path and name
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);
    
    return iaxios.post(`/core/space/${installId}/files/upload`, formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
}

export const createSpaceFolder = async (installId: number, name: string, path: string = '') => {
    return iaxios.post(`/core/space/${installId}/files/folder`, {
        name,
        path,
    });
}

// Presigned Upload API
export interface PresignedUploadResponse {
    presigned_token: string;
    upload_url: string;
    expiry: number;
}

export const createPresignedUploadURL = async (installId: number, fileName: string, path: string = '', expiry?: number) => {
    return iaxios.post<PresignedUploadResponse>(`/core/space/${installId}/files/presigned`, {
        file_name: fileName,
        path,
        expiry,
    });
}

export const uploadFileWithPresignedToken = async (presignedKey: string, file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    
    return axios.post(`/zz/file/upload-presigned?presigned-key=${presignedKey}`, formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
}

// Capability Types API
export interface CapabilityOptionField {
    name: string;
    key: string;
    description: string;
    type: string; // text, number, date, api_key, boolean, select, multi_select, textarea
    default: string;
    options: string[];
    required: boolean;
}

export interface CapabilityDefinition {
    name: string;
    icon: string;
    option_fields: CapabilityOptionField[];
}

export const listCapabilityTypes = async () => {
    return iaxios.get<CapabilityDefinition[]>(`/core/capability/types`);
}

// Space Capabilities API
export interface SpaceCapability {
    id: number;
    name: string;
    capability_type: string;
    install_id: number;
    space_id: number;
    options: string; // JSON string
    extrameta: string; // JSON string
}

export const listSpaceCapabilities = async (installId: number, spaceId?: number, capabilityType?: string) => {
    return iaxios.get<SpaceCapability[]>(`/core/space/${installId}/capabilities`, {
        params: {
            ...(spaceId !== undefined && { space_id: spaceId }),
            ...(capabilityType && { capability_type: capabilityType }),
        },
    });
}

export const getSpaceCapability = async (installId: number, capabilityId: number) => {
    return iaxios.get<SpaceCapability>(`/core/space/${installId}/capabilities/${capabilityId}`);
}

export const createSpaceCapability = async (installId: number, data: {
    name: string;
    capability_type: string;
    space_id?: number; // 0 or omitted for package-level, >0 for space-level
    options?: any; // Will be JSON stringified
    extrameta?: any; // Will be JSON stringified
}) => {
    return iaxios.post<SpaceCapability>(`/core/space/${installId}/capabilities`, data);
}

export const updateSpaceCapability = async (installId: number, capabilityId: number, data: Partial<SpaceCapability>) => {
    return iaxios.put<SpaceCapability>(`/core/space/${installId}/capabilities/${capabilityId}`, data);
}

export const deleteSpaceCapability = async (installId: number, capabilityId: number) => {
    return iaxios.delete<void>(`/core/space/${installId}/capabilities/${capabilityId}`);
}

// Space Users API
export interface SpaceUser {
    id: number;
    user_id: number;
    install_id: number;
    space_id: number;
    scope: string;
    token: string;
    extrameta: string;
}

export const listSpaceUsers = async (installId: number, spaceId?: number, userId?: number, scope?: string) => {
    return iaxios.get<SpaceUser[]>(`/core/space/${installId}/users`, {
        params: {
            ...(spaceId !== undefined && { space_id: spaceId }),
            ...(userId !== undefined && { user_id: userId }),
            ...(scope && { scope }),
        },
    });
}

export const getSpaceUser = async (installId: number, spaceUserId: number) => {
    return iaxios.get<SpaceUser>(`/core/space/${installId}/users/${spaceUserId}`);
}

export const createSpaceUser = async (installId: number, data: {
    user_id: number;
    space_id?: number; // 0 or omitted for package-level, >0 for space-level
    scope?: string;
    token?: string;
    extrameta?: string;
}) => {
    return iaxios.post<SpaceUser>(`/core/space/${installId}/users`, data);
}

export const updateSpaceUser = async (installId: number, spaceUserId: number, data: Partial<SpaceUser>) => {
    return iaxios.put<SpaceUser>(`/core/space/${installId}/users/${spaceUserId}`, data);
}

export const deleteSpaceUser = async (installId: number, spaceUserId: number) => {
    return iaxios.delete<void>(`/core/space/${installId}/users/${spaceUserId}`);
}

// Event Subscriptions API
export interface EventSubscription {
    id: number;
    install_id: number;
    space_id: number;
    event_key: string;
    target_type: string; // push, email, sms, webhook, script
    target_endpoint: string;
    target_options: string; // JSON string
    target_code: string;
    rules: string; // JSON string
    transform: string; // JSON string
    delay_start: number;
    retry_delay: number;
    max_retries: number;
    extrameta: string; // JSON string
    created_by: number;
    disabled: boolean;
    created_at?: string;
    updated_at?: string;
}

export const listEventSubscriptions = async (installId: number, spaceId?: number, eventKey?: string) => {
    return iaxios.get<EventSubscription[]>(`/core/space/${installId}/events`, {
        params: {
            ...(spaceId !== undefined && { space_id: spaceId }),
            ...(eventKey && { event_key: eventKey }),
        },
    });
}

export const getEventSubscription = async (installId: number, subscriptionId: number) => {
    return iaxios.get<EventSubscription>(`/core/space/${installId}/events/${subscriptionId}`);
}

export const createEventSubscription = async (installId: number, data: {
    event_key: string;
    target_type: string;
    space_id?: number; // 0 or omitted for package-level, >0 for space-level
    target_endpoint?: string;
    target_options?: any; // Will be JSON stringified
    target_code?: string;
    rules?: any; // Will be JSON stringified
    transform?: any; // Will be JSON stringified
    delay_start?: number;
    retry_delay?: number;
    max_retries?: number;
    extrameta?: any; // Will be JSON stringified
    disabled?: boolean;
}) => {
    // Stringify JSON fields
    const payload: any = {
        event_key: data.event_key,
        target_type: data.target_type,
        ...(data.space_id !== undefined && { space_id: data.space_id }),
        ...(data.target_endpoint && { target_endpoint: data.target_endpoint }),
        ...(data.target_options && { target_options: typeof data.target_options === 'string' ? data.target_options : JSON.stringify(data.target_options) }),
        ...(data.target_code && { target_code: data.target_code }),
        ...(data.rules && { rules: typeof data.rules === 'string' ? data.rules : JSON.stringify(data.rules) }),
        ...(data.transform && { transform: typeof data.transform === 'string' ? data.transform : JSON.stringify(data.transform) }),
        ...(data.delay_start !== undefined && { delay_start: data.delay_start }),
        ...(data.retry_delay !== undefined && { retry_delay: data.retry_delay }),
        ...(data.max_retries !== undefined && { max_retries: data.max_retries }),
        ...(data.extrameta && { extrameta: typeof data.extrameta === 'string' ? data.extrameta : JSON.stringify(data.extrameta) }),
        ...(data.disabled !== undefined && { disabled: data.disabled }),
    };
    
    return iaxios.post<EventSubscription>(`/core/space/${installId}/events`, payload);
}

export const updateEventSubscription = async (installId: number, subscriptionId: number, data: Partial<EventSubscription>) => {
    // Stringify JSON fields if they're objects
    const payload: any = {};
    if (data.event_key !== undefined) payload.event_key = data.event_key;
    if (data.target_type !== undefined) payload.target_type = data.target_type;
    if (data.space_id !== undefined) payload.space_id = data.space_id;
    if (data.target_endpoint !== undefined) payload.target_endpoint = data.target_endpoint;
    if (data.target_options !== undefined) payload.target_options = typeof data.target_options === 'string' ? data.target_options : JSON.stringify(data.target_options);
    if (data.target_code !== undefined) payload.target_code = data.target_code;
    if (data.rules !== undefined) payload.rules = typeof data.rules === 'string' ? data.rules : JSON.stringify(data.rules);
    if (data.transform !== undefined) payload.transform = typeof data.transform === 'string' ? data.transform : JSON.stringify(data.transform);
    if (data.delay_start !== undefined) payload.delay_start = data.delay_start;
    if (data.retry_delay !== undefined) payload.retry_delay = data.retry_delay;
    if (data.max_retries !== undefined) payload.max_retries = data.max_retries;
    if (data.extrameta !== undefined) payload.extrameta = typeof data.extrameta === 'string' ? data.extrameta : JSON.stringify(data.extrameta);
    if (data.disabled !== undefined) payload.disabled = data.disabled;
    
    return iaxios.put<EventSubscription>(`/core/space/${installId}/events/${subscriptionId}`, payload);
}

export const deleteEventSubscription = async (installId: number, subscriptionId: number) => {
    return iaxios.delete<void>(`/core/space/${installId}/events/${subscriptionId}`);
}

// User Messages API
export interface UserMessage {
    id: number;
    title: string;
    is_read: boolean;
    type: string;
    contents: string;
    to_user: number;
    from_user_id: number;
    from_space_id: number;
    callback_token: string;
    warn_level: number;
    created_at?: string;
}

export const listUserMessages = async (afterId?: number, limit: number = 100) => {
    const params: any = { limit };
    if (afterId !== undefined && afterId !== null) {
        params.after_id = afterId;
    }
    return iaxios.get<UserMessage[]>(`/core/user/messages`, { params });
}

export const queryNewMessages = async () => {
    return iaxios.get<UserMessage[]>(`/core/user/messages/new`);
}

export const queryMessageHistory = async (limit: number = 100) => {
    return iaxios.get<UserMessage[]>(`/core/user/messages/history`, {
        params: {
            limit,
        },
    });
}

export const getUserMessage = async (id: number) => {
    return iaxios.get<UserMessage>(`/core/user/messages/${id}`);
}

export const sendUserMessage = async (data: {
    title: string;
    type: string;
    contents: string;
    to_user: number;
    from_space_id?: number;
    callback_token?: string;
    warn_level?: number;
}) => {
    return iaxios.post<UserMessage>(`/core/user/messages`, data);
}

export const updateUserMessage = async (id: number, data: Partial<UserMessage>) => {
    return iaxios.put<UserMessage>(`/core/user/messages/${id}`, data);
}

export const deleteUserMessage = async (id: number) => {
    return iaxios.delete<void>(`/core/user/messages/${id}`);
}

export const setMessageAsRead = async (id: number) => {
    return iaxios.post<{ message: string }>(`/core/user/messages/${id}/read`);
}

export const setAllMessagesAsRead = async () => {
    return iaxios.post<{ message: string; read_head: number }>(`/core/user/messages/read-all`);
}