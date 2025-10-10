import axios, { AxiosInstance } from "axios";
import { getLoginData } from "./utils";


let iaxios: AxiosInstance = axios.create({
    baseURL: "/zz/api",
});

export const initHttpClient = () => {

    const data = getLoginData();

    const headers: Record<string, string> = {
        "Content-Type": "application/json",
        "X-Overhead": "Aaron Swartz",
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

export const installPackageEmbed = async (name: string) => {
    return iaxios.post<{ package_id: number }>(`/core/package/install/embed`, { name });
}

export const deletePackage = async (id: number) => {
    return iaxios.delete<void>(`/core/package/${id}`);
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

export const listEPackages = async () => {
    return iaxios.get<EPackage[]>(`/core/package/list`);
}


export interface Package {
    id: number;
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
    package_id: number;
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


export interface SpaceInfo {
    id: number;
    namespace_key: string;
    owns_namespace: boolean;
    package_name: string;
    package_info: string;
}

export const getSpaceInfo = async (space_key: string) => {
    return iaxios.get<SpaceInfo>(`/core/engine/space_info/${space_key}`);
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
    return iaxios.get<PackageFile[]>(`/core/package/${packageId}/files`, {
        params: {
            path,
        },
    });
}

export const getPackageFile = async (packageId: number, fileId: number) => {
    return iaxios.get<PackageFile>(`/core/package/${packageId}/files/${fileId}`);
}

export const downloadPackageFile = async (packageId: number, fileId: number) => {
    return iaxios.get(`/core/package/${packageId}/files/${fileId}/download`, {
        responseType: 'blob'
    });
}

export const deletePackageFile = async (packageId: number, fileId: number) => {
    return iaxios.delete(`/core/package/${packageId}/files/${fileId}`);
}

export const uploadPackageFile = async (packageId: number, file: File, path: string = '') => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);
    
    return iaxios.post(`/core/package/${packageId}/files/upload`, formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
}

// Space KV API
export interface SpaceKV {
    id: number;
    key: string;
    group_name: string;
    value: string;
    space_id: number;
    tag1: string;
    tag2: string;
    tag3: string;
}

export const listSpaceKV = async (spaceId: number) => {
    return iaxios.get<SpaceKV[]>(`/core/space/${spaceId}/kv`);
}

export const getSpaceKV = async (spaceId: number, id: number) => {
    return iaxios.get<SpaceKV>(`/core/space/${spaceId}/kv/${id}`);
}

export const createSpaceKV = async (spaceId: number, data: {
    key: string;
    group_name: string;
    value: string;
    tag1?: string;
    tag2?: string;
    tag3?: string;
}) => {
    return iaxios.post<SpaceKV>(`/core/space/${spaceId}/kv`, data);
}

export const updateSpaceKV = async (spaceId: number, id: number, data: {
    key?: string;
    group_name?: string;
    value?: string;
    tag1?: string;
    tag2?: string;
    tag3?: string;
}) => {
    return iaxios.put<SpaceKV>(`/core/space/${spaceId}/kv/${id}`, data);
}

export const deleteSpaceKV = async (spaceId: number, id: number) => {
    return iaxios.delete<void>(`/core/space/${spaceId}/kv/${id}`);
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

export const listSpaceFiles = async (spaceId: number, path: string = '') => {
    return iaxios.get<SpaceFile[]>(`/core/space/${spaceId}/files`, {
        params: {
            path,
        },
    });
}

export const getSpaceFile = async (spaceId: number, fileId: number) => {
    return iaxios.get<SpaceFile>(`/core/space/${spaceId}/files/${fileId}`);
}

export const downloadSpaceFile = async (spaceId: number, fileId: number) => {
    return iaxios.get(`/core/space/${spaceId}/files/${fileId}/download`, {
        responseType: 'blob'
    });
}

export const deleteSpaceFile = async (spaceId: number, fileId: number) => {
    return iaxios.delete(`/core/space/${spaceId}/files/${fileId}`);
}

export const uploadSpaceFile = async (spaceId: number, file: File, path: string = '') => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);
    
    return iaxios.post(`/core/space/${spaceId}/files/upload`, formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
}

export const createSpaceFolder = async (spaceId: number, name: string, path: string = '') => {
    return iaxios.post(`/core/space/${spaceId}/files/folder`, {
        name,
        path,
    });
}