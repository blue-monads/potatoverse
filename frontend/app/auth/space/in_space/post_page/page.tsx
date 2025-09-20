"use client";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

export default function InSpacePostPage() {
    const params = useSearchParams();

    useEffect(() => {
        if (typeof window === 'undefined') {
            return;
        }

        const redirect_back_url = params.get('redirect_back_url');
        if (!redirect_back_url) {
            console.log("@redirect_back_url is not set");
            return;
        }

        
        window.location.href = redirect_back_url;
    
        
        
    }, [params]);


    return (<>
        <div className="flex flex-col items-center justify-center h-screen">
            <div>post Loading...</div>
        </div>
    </>)
}
