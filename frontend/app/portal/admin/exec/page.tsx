"use client";
import { deriveHost } from '@/lib/api';
import { Loader2 } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import React, { useEffect, useRef, useState } from 'react';

// /portal/admin/exec?nskey=test&space_id=1

const buildIframeSrc = (nskey: string, host: string) => {

    let src = `/zz/space/${nskey}`;

    if (host) {
        const origin = window.location.origin;
        const isSecure = origin.startsWith("https//");
        const hasPort = origin.includes(":");
        const port = hasPort ? origin.split(":").at(-1) : "";

        src = `${isSecure ? "https://" : "http://"}${host}${hasPort ? ":" + port : ""}/zz/space/${nskey}`;
    } 

    return src;
}


export default function Page() {
    const searchParams = useSearchParams();
    const nskey = searchParams.get('nskey');
    const space_id = searchParams.get('space_id');
    const load_page = searchParams.get('load_page');
    const [isLoading, setIsLoading] = useState(true);
    const iframeRef = useRef<HTMLIFrameElement>(null);
    const [iframeSrc, setIframeSrc] = useState('');

    const loadHost = async () => {
        if (!nskey) {
            return;
        }

        setIsLoading(true);        

        try {
            const resp = await deriveHost(nskey, space_id || undefined);
            if (resp.status !== 200) {
                console.error("failed to derive host");
                return;
            }
        
            setIframeSrc(buildIframeSrc(nskey, resp.data.host));

            setTimeout(() => {
                setIsLoading(false);
            }, 300)


        } catch (error) {
            console.error(error);
        }


    }



    useEffect(() => {
        if (!nskey || !space_id) {
            return;
        }

        loadHost();
        
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

                {iframeSrc && (<>
                    <iframe

                        ref={iframeRef}
                        src={load_page ? `${iframeSrc}/${load_page}` : iframeSrc}
                        // src={`/zz/test_page.html`}
                        className={isLoading ? 'h-[1px] w-[1px]' : 'w-full h-full flex-grow'}
                    ></iframe>
                </>)}



            </div>
        </div>
    );
}