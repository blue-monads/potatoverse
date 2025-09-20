"use client";
import { UserInfo } from "@/hooks";
import { getLoginData, getSpaceInfo, SpaceInfo } from "@/lib";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

export default function InHostPage() {
    return (<>
        <div className="flex flex-col items-center justify-center h-screen">
            <InSpaceAuthorizerWrapper />
        </div>
    </>)
}



const InSpaceAuthorizerWrapper = () => {
    const [spaceInfo, setSpaceInfo] = useState<SpaceInfo | null>(null);
    const [mode, setMode] = useState<"loading" | "error" | "success">('loading');
    const [error, setError] = useState<string | null>(null);
    const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
    const [isAuthenticated, setIsAuthenticated] = useState(false);

    const params = useSearchParams();


    useEffect(() => {
        if (typeof window === 'undefined') {
            return;
        }


        const fetchSpaceInfo = async () => {
            setMode('loading');
            setError(null);

            const redirect_back_url = params.get('redirect_back_url');
            if (!redirect_back_url) {
                setError("redirect_back_url is not set");
                setMode('error');
                return;
            }

            const redirect_back_url_url = new URL(redirect_back_url);

            if (!redirect_back_url_url.pathname.startsWith('/z/space/')) {
                setError("redirect_back_url is not a valid redirect_back_url");
                setMode('error');
                return;
            }



            const url = new URL(redirect_back_url);
            const nspace_key = url.pathname.split('/').pop();
            if (!nspace_key) {
                setError("extract space_key failed");
                setMode('error');
                return;
            }


            const resp = await getSpaceInfo(nspace_key);
            if (resp.status === 200) {
                setSpaceInfo(resp.data);
                setMode('success');
            } else {
                setError("failed to get space info");
                setMode('error');
            }

        }

        fetchSpaceInfo();

    }, [params]);


    useEffect(() => {
        if (typeof window === 'undefined') {
            return;
        }

        const checkLoginData = () => {
            const data = getLoginData();
            if (data?.userInfo) {
                setUserInfo(data.userInfo);
                setIsAuthenticated(true);
            } else {
                setIsAuthenticated(false);
            }
        }

        checkLoginData();
    }, [])



    return (<>

        {mode === "loading" && (<>
            <div>Loading...</div>
        </>)}

        {mode === "error" && (<>
            <div>Error: {error}</div>
        </>)}

        {mode === "success" && (<>

            {isAuthenticated && (<>
                <NotAuthorizedPromptCard
                    onLogin={() => {
                        const redirect_back_url = params.get('redirect_back_url');
                        if (!redirect_back_url) {
                            console.log("@redirect_back_url is not set");
                            return;
                        }

                        window.sessionStorage.setItem('redirect_back_url', redirect_back_url);
                        const loginPageUrl = new URL('/z/pages/auth/login', window.location.origin);

                        const finalRedirectBackUrl = new URL("/z/pages/auth/space/in_host", window.location.origin);
                        finalRedirectBackUrl.searchParams.set('redirect_back_url', redirect_back_url);

                        loginPageUrl.searchParams.set('after_login_redirect_back_url', finalRedirectBackUrl.toString());
                        window.location.href = loginPageUrl.toString();

                    }}
                    onDeny={() => { }}
                />


            </>)}

            {!isAuthenticated && (<>


                <AuthorizePromptCard
                    spaceInfo={spaceInfo!}
                    onAuthorize={() => { }}
                    onDeny={() => { }}
                    onChangeAccount={() => { }}
                    loggedInUser={userInfo?.name!}
                />



            </>)}



        </>)}


    </>)


}


interface PromptCardProps {
    spaceInfo: SpaceInfo;
    onAuthorize: () => void;
    onDeny: () => void;
    onChangeAccount: () => void;
    loggedInUser: string;
}


const AuthorizePromptCard = (props: PromptCardProps) => {
    return (
        <div className="flex items-center justify-center min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-8 w-full max-w-md flex flex-col gap-4">
                <div className="flex mb-6 space-x-4 items-center justify-center">
                    <img src="/z/pages/logo.png" alt="Turnix Logo" className="w-10 h-10" />
                </div>

                <h6 className="h4 text-base">
                    Do you want to authorize this space?
                </h6>

                <div className="flex justify-center items-center">
                    <span className="font-light rounded-md bg-gray-100 p-2">{props.spaceInfo.package_name}</span>
                </div>



                <div className="space-y-3 mb-6">
                    <button
                        onClick={props.onAuthorize}
                        className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition duration-150 ease-in-out"
                    >
                        Authorize
                    </button>
                    <button
                        onClick={props.onDeny}
                        className="w-full bg-gray-200 text-gray-800 py-2 px-4 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-2 transition duration-150 ease-in-out dark:bg-gray-700 dark:text-white dark:hover:bg-gray-600"
                    >
                        Deny
                    </button>
                </div>

                <p className="text-center text-sm text-gray-500 dark:text-gray-400">
                    Logged in as <span className="font-medium">{props.loggedInUser}</span>.{" "}
                    <button onClick={props.onChangeAccount} className="text-blue-600 hover:underline dark:text-blue-400">
                        Change account
                    </button>
                </p>
            </div>
        </div>
    );
};

interface NotAuthorizedPromptCardProps {
    onLogin: () => void;
    onDeny: () => void;
}

const NotAuthorizedPromptCard = (props: NotAuthorizedPromptCardProps) => {

    return (
        <div className="flex items-center justify-center min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-8 w-full max-w-md flex flex-col gap-4">
                <div className="flex mb-6 space-x-4 items-center justify-center">
                    <img src="/z/pages/logo.png" alt="Turnix Logo" className="w-10 h-10" />
                </div>
                <h6 className="h4 text-base">
                    You are not logged in, Please login first to authorize this space.
                </h6>

                <div className="space-y-3 mb-6">
                    <button
                        onClick={props.onLogin}
                        className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition duration-150 ease-in-out"
                    >
                        Login
                    </button>
                    <button
                        onClick={props.onDeny}
                        className="w-full bg-gray-200 text-gray-800 py-2 px-4 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-2 transition duration-150 ease-in-out dark:bg-gray-700 dark:text-white dark:hover:bg-gray-600"
                    >
                        Deny
                    </button>
                </div>


            </div>


        </div>
    )

}