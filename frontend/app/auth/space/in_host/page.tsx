"use client";
import { UserInfo } from "@/hooks";
import { authorizeSpace, getLoginData, getSpaceInfo, SpaceInfo } from "@/lib";
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
    const [mode, setMode] = useState<"loading" | "error" | "space_info_loaded" | "space_token_loaded">('loading');
    const [error, setError] = useState<string | null>(null);
    const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [spaceToken, setSpaceToken] = useState<string | null>(null);

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

            if (!redirect_back_url_url.pathname.startsWith('/zz/space/')) {
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

            try {
                const resp = await getSpaceInfo(nspace_key, redirect_back_url_url.hostname);
                if (resp.status === 200) {
                    setSpaceInfo(resp.data);
                    setMode('space_info_loaded');
                } else {
                    setError("failed to get space info");
                    setMode('error');
                }
    
                
            } catch (error) {
                console.error(error);
                setError("failed to get space info: " + error);
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


    const redirectWithoutSpaceToken = () => {
        console.log("@redirectWithoutSpaceToken", params.toString());

        const redirect_back_url = params.get('redirect_back_url');
        if (!redirect_back_url) {
            console.log("@redirect_back_url is not set");
            return;
        }

        const actual_page = params.get('actual_page');

        const redirect_back_url_url = new URL(redirect_back_url);
        if (actual_page) {
            redirect_back_url_url.searchParams.set('actual_page', actual_page);
        }
        redirect_back_url_url.searchParams.set("deny_space_token", "true");
        window.location.href = redirect_back_url_url.toString();
    }


    const redirrectToLoginPage = () => {
        console.log("@redirrectToLoginPage", params.toString());

        const redirect_back_url = params.get('redirect_back_url');
        if (!redirect_back_url) {
            console.log("@redirect_back_url is not set");
            return;
        }
        
        const actual_page = params.get('actual_page');


        window.sessionStorage.setItem('redirect_back_url', redirect_back_url);
        const loginPageUrl = new URL('/zz/pages/auth/login', window.location.origin);

        const finalRedirectBackUrl = new URL("/zz/pages/auth/space/in_host", window.location.origin);
        finalRedirectBackUrl.searchParams.set('redirect_back_url', redirect_back_url);

        if (actual_page) {
            finalRedirectBackUrl.searchParams.set('actual_page', actual_page);
        }

        console.log("@finalRedirectBackUrl", finalRedirectBackUrl.toString());

        loginPageUrl.searchParams.set('after_login_redirect_back_url', finalRedirectBackUrl.toString());
        window.location.href = loginPageUrl.toString();

    }

    const getSpaceToken = async () => {
        if (!spaceInfo) {
            setError("spaceInfo is not set");
            setMode('error');
            return;
        }
        
        try {

            setMode('loading');


            const resp = await authorizeSpace(spaceInfo.namespace_key, spaceInfo.id);
            if (resp.status === 200) {
                setSpaceToken(resp.data.token);

                // localStorage.setItem(`${spaceInfo.namespace_key}_space_token`, resp.data.token);

                setMode('space_token_loaded');
            } else {
                setError("failed to authorize space");
                setMode('error');
            }


        } catch (error) {
            console.error(error);
            setError("failed to authorize space: " + error);
            setMode('error');
        }


    }



    return (<>

        {mode === "loading" && (<>
            <div>Loading...</div>
        </>)}

        {mode === "error" && (<>
            <div>Error: {error}</div>
        </>)}

        {mode === "space_info_loaded" && (<>

            {!isAuthenticated && (<>
                <NotAuthorizedPromptCard
                    onLogin={redirrectToLoginPage}
                    onDeny={redirectWithoutSpaceToken}
                />


            </>)}

            {isAuthenticated && (<>


                <AuthorizePromptCard
                    spaceInfo={spaceInfo!}
                    onAuthorize={() => {
                        getSpaceToken();
                    }}
                    onDeny={redirectWithoutSpaceToken}
                    onChangeAccount={redirrectToLoginPage}
                    loggedInUser={userInfo?.name!}
                />



            </>)}



        </>)}

        {mode === "space_token_loaded" && (<>
            <div className="flex items-center justify-center min-h-[500px] bg-gray-100 p-4">
                <div className="bg-white rounded-lg shadow-xl p-8 w-full max-w-md flex flex-col gap-4">
                    <div className="flex mb-6 space-x-4 items-center justify-center">
                        <img src="/zz/pages/logo.png" alt="Turnix Logo" className="w-10 h-10" />
                    </div>
                    
                    <h6 className="h4 text-base">
                        Authorized successfully
                    </h6>

                    <div className="flex justify-center items-center">
                        <button onClick={() => {

                            console.log("@in_host/post_page", params.toString());


                            const redirect_back_url = params.get('redirect_back_url');
                            if (!redirect_back_url) {
                                console.log("@redirect_back_url is not set");
                                return;
                            }

                            const actual_page = params.get('actual_page');

                            const redirect_back_url_url = new URL("/zz/pages/auth/space/in_space/post_page", redirect_back_url);
                            //const redirect_back_url_url = new URL("/zz/pages/auth/space/in_space/post_page", window.origin);
                            redirect_back_url_url.searchParams.set("redirect_back_url", redirect_back_url);
                            redirect_back_url_url.searchParams.set("space_token", spaceToken!);
                            redirect_back_url_url.searchParams.set("nskey", spaceInfo!.namespace_key);

                            if (actual_page) {
                                redirect_back_url_url.searchParams.set('actual_page', actual_page);
                            }




                            window.location.href = redirect_back_url_url.toString();
                        }}>
                            Redirect to space
                        </button>

                    </div>
                </div>

            </div>

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
        <div className="flex items-center justify-center h-[500px] bg-gray-100 p-4">
            <div className="bg-white rounded-lg shadow-xl p-8 w-full max-w-md flex flex-col gap-4">
                <div className="flex mb-6 space-x-4 items-center justify-center">
                    <img src="/zz/pages/logo.png" alt="Turnix Logo" className="w-10 h-10" />
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
                        className="w-full bg-gray-200 text-gray-800 py-2 px-4 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-2 transition duration-150 ease-in-out"
                    >
                        Deny
                    </button>
                </div>

                <p className="text-center text-sm text-gray-500">
                    Logged in as <span className="font-medium">{props.loggedInUser}</span>.{" "}
                    <button onClick={props.onChangeAccount} className="text-blue-600 hover:underline">
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
        <div className="flex items-center justify-center h-[500px] bg-gray-100 p-4">
            <div className="bg-white rounded-lg shadow-xl p-8 w-full max-w-md flex flex-col gap-4">
                <div className="flex mb-6 space-x-4 items-center justify-center">
                    <img src="/zz/pages/logo.png" alt="Turnix Logo" className="w-10 h-10" />
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
                        className="w-full bg-gray-200 text-gray-800 py-2 px-4 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-2 transition duration-150 ease-in-out"
                    >
                        Deny
                    </button>
                </div>


            </div>


        </div>
    )

}