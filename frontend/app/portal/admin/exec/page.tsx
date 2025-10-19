"use client";
import { Loader2 } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import React, { useEffect, useRef, useState } from 'react';

// /portal/admin/exec?nskey=test&space_id=1

const buildIframeSrc = (nskey: string, space_id: string) => {
    const attrs = (window as any).__potato_attrs__ || {};
    let host = attrs.site_host || '';
    let src = '';
    if (host) {
        const origin = window.location.origin;
        const isSecure = origin.startsWith("https//");
        const hasPort = origin.includes(":");
        const port = hasPort ? origin.split(":").at(-1) : "";
        if (host === "*") {
            host = `*.${window.location.host.split(":")[0]}`;
        }

        if (host.includes('*')) {
            src = `${isSecure ? "https://" : "http://"}${host.replace('*', "s-" + space_id) }${hasPort ? ":" + port : ""}/zz/space/${nskey}`;
        } else {
            src = `${isSecure ? "https://" : "http://"}${host}${hasPort ? ":" + port : ""}/zz/space/${nskey}`;
        }
    } else {
        src = `/zz/space/${nskey}`;
    }

    return src;
}


export default function Page() {
    const searchParams = useSearchParams();
    const nskey = searchParams.get('nskey');
    const space_id = searchParams.get('space_id');
    const [isLoading, setIsLoading] = useState(true);
    const iframeRef = useRef<HTMLIFrameElement>(null);
    const [iframeSrc, setIframeSrc] = useState('');

    useEffect(() => {
        if (!nskey || !space_id) {
            return;
        }


        const timer = setTimeout(() => {
            setIsLoading(false);
        }, 300);

        setIframeSrc(buildIframeSrc(nskey, space_id));


        return () => clearTimeout(timer);
    }, [nskey, space_id]);

    console.log(iframeSrc);

    return (
        <div className='p-1'>
            <div className='p-1 rounded-md w-full min-h-[99vh] border border-primary-100 flex flex-col'>

               {isLoading && (
                    <div className='flex items-center justify-center h-full'>
                        <Loader2 className='w-12 h-12 animate-spin my-20' />
                    </div>
                )} 

                {iframeRef && (<>
                    <iframe

                        ref={iframeRef}
                        src={iframeSrc}
                        // src={`/zz/test_page.html`}
                        className={isLoading ? 'h-[1px] w-[1px]' : 'w-full h-full flex-grow'}
                    ></iframe>                
                </>)}



            </div>
        </div>
    );
}