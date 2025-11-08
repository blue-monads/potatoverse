"use client";
import { UserInfo } from "@/hooks";
import { getLoginData, getSpaceInfo, SpaceInfo } from "@/lib";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

const extractDomainSPrefixRegex = /^(https?:\/\/)s-\d+\./;

export default function InSpacePrePage() {
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

        
        // FIXME => in future if space is running on different domain, we need to change this to use the correct domain        
        // *.coolapps.com |>  xyzapp.coolapps.com -> xyzapp.coolapps.com/zz/pages/auth/space/in_space/pre_page -> coolapps.com/zz/pages/auth/space/in_host?redirect_back_url=/zz/space/xyzapp

        // s-123.example.com -> example.com

        let origin = window.location.origin;
        if (extractDomainSPrefixRegex.test(origin)) {
            origin = origin.replace(extractDomainSPrefixRegex, '$1');
        }

        const authorizerPageUrl = new URL('/zz/pages/auth/space/in_host', origin);
        const finalRedirectBackUrl = new URL(redirect_back_url, window.location.origin);
        authorizerPageUrl.searchParams.set('redirect_back_url', finalRedirectBackUrl.toString());

        window.location.href = authorizerPageUrl.toString();

        
        
    }, [params]);


    return (<>
        <div className="flex flex-col items-center justify-center h-screen">
            <div>Pre Loading...</div>
        </div>
    </>)
}
