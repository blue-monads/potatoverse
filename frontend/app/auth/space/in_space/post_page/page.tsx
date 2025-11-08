"use client";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

export default function InSpacePostPage() {
    const params = useSearchParams();

    useEffect(() => {
        if (typeof window === 'undefined') {
            return;
        }

        console.log("@POSTPAGE", params.toString())

        const redirect_back_url = params.get('redirect_back_url');
        const space_token = params.get('space_token');
        let nskey = params.get('nskey');
        if (!nskey) {
            console.log("@post/3");
            const mayBenskey = redirect_back_url?.split('/').pop() 
            if (mayBenskey) {
                console.log("@post/4");

                nskey = mayBenskey;
            }
            console.log("@post/5");

        }

        console.log("@post/6");


        if (nskey && space_token) {
            console.log("@post/7", nskey, space_token);

            localStorage.setItem(`${nskey}_space_token`, space_token);

            console.log("@post/7.1");

        }

        console.log("@post/8");


        if (!nskey) {
            console.log("@post/9");

            console.log("@nskey is not set");
            return;
        }

        console.log("@post/10");

        
        if (!redirect_back_url) {
            console.log("@post/11");

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
