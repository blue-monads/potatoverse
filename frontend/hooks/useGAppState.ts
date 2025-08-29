import { getAccessToken } from "@/lib";
import { useEffect, useState } from "react";


export const useGAppState = () => {    
    const [isLoading, setIsLoading] = useState(false);
    const [isAuthenticated, setIsAuthenticated] = useState(false);

    const checkToken = () => {
        const accessToken = getAccessToken();
        if (accessToken) {
            setIsAuthenticated(true);
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
    }
}

export type Handle = ReturnType<typeof useGAppState>;