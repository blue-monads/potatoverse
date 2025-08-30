import { getLoginData, initHttpClient } from "@/lib";
import { useEffect, useState } from "react";


export interface UserInfo {
    id: number;
    name: string;
    username: string;
    email: string;
}

export const useGAppState = () => {    
    const [isLoading, setIsLoading] = useState(false);
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [userInfo, setUserInfo] = useState<UserInfo | null>(null);

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
        setIsLoading(false);        
    }


    useEffect(() => {
        checkToken();
    }, []);


    return {
        isLoading,
        isAuthenticated,
        checkToken,
        userInfo,
    }
}

export type Handle = ReturnType<typeof useGAppState>;