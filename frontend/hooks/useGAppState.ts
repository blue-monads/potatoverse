import { getLoginData, initHttpClient, removeLoginData, saveLoginData } from "@/lib";
import { useEffect, useState } from "react";
import { useGModal, ModalHandle } from "./modal/useGModal";

export interface UserInfo {
    id: number;
    name: string;
    username: string;
    email: string;
}

export const useGAppState = () => {    
    const [loaded, setLoaded] = useState(false);
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
    const modal = useGModal();

    console.log("userInfo", userInfo);

    const checkToken = () => {
        const data = getLoginData();
        console.log("@getLoginData", data);
        if (data?.accessToken) {

            setUserInfo(data.userInfo);
            setIsAuthenticated(true);
            initHttpClient();
        } else {
            setIsAuthenticated(false);
        }        
        setLoaded(true);        
    }

    const logOut = () => {
        removeLoginData();
        setUserInfo(null);
        setIsAuthenticated(false);
    }

    const logIn = (token: string, userInfo: UserInfo) => {
        saveLoginData(token, userInfo);
        checkToken();
    }

    useEffect(() => {
        checkToken();
    }, []);

    return {
        loaded,
        isAuthenticated,
        checkToken,
        logOut,
        logIn,
        userInfo,
        modal,
    }
}

export type Handle = ReturnType<typeof useGAppState>;