"use client";
import { deriveHost } from '@/lib/api';
import { Loader2 } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import React, { useEffect, useRef, useState } from 'react';
import { deriveHostAndIframeSrc } from './hostSrc';

// /portal/admin/exec?nskey=test&space_id=1


export default function Page() {
    const searchParams = useSearchParams();
    const nskey = searchParams.get('nskey');
    const space_id = searchParams.get('space_id');
    const load_page = searchParams.get('load_page');
    const [isLoading, setIsLoading] = useState(true);
    const iframeRef = useRef<HTMLIFrameElement>(null);
    const [iframeSrc, setIframeSrc] = useState('');
    const [errmsg, setErrmsg] = useState<string>();

    const loadHost = async () => {
        if (!nskey || !space_id) {
            setErrmsg("invalid link or unknown error")
            return;
        }

        
        try {

            setIsLoading(true);


            const startTime = Date.now();
            const hostsrc = await deriveHostAndIframeSrc(nskey, space_id)
            const endTime = Date.now();

            const duration = endTime - startTime;
            console.log(`deriveHost took ${duration}ms`);
            if (!hostsrc) {
                setErrmsg("Error building src")
                return
            }


            setIframeSrc(hostsrc);

            setTimeout(() => {
                setIsLoading(false);
            }, Math.min(Math.max(200, duration), 10000))

        } catch (error) {
            console.error(error);
        }


    }

    console.log("@", {
        isLoading,
        iframeSrc,
        errmsg
    })



    useEffect(() => {
        
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

                {errmsg && (<>
                    <div className='flex items-center justify-center h-full'>
                        <p className="text-red-500">{errmsg}</p>
                    </div>
                </>)}

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