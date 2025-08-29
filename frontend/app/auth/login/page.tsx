"use client"
import Image from "next/image";
import WithLoginLayout from "./WithLoginLayout";
import { useState } from "react";
import { initHttpClient, login } from "@/lib/api";
import { saveAccessToken } from "@/lib";
import { useRouter } from "next/navigation";


export default function Page() {

    const [email, setEmail] = useState<string>("demo@example.com");
    const [password, setPassword] = useState<string>("demogodTheGreat_123");
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string>("");

    const router = useRouter();

    const handleSubmit = async (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();
        setLoading(true);
        try {
            const res = await login(email, password);
            if (res.status !== 200) {
                setError("An unknown error occurred");
                return;
            }

            const token = res.data.access_token;
            saveAccessToken(token);
            initHttpClient();

            router.push("/z/pages/admin");



        } catch (err) {
            setError(err instanceof Error ? err.message : "An unknown error occurred"   );
        } finally {
            setLoading(false);
        }
    }


    return (<>
        <WithLoginLayout    >
            <div className="w-full max-w-md space-y-8 px-4 bg-white text-gray-600 sm:px-0">
                <div className="">
                    <Image
                        className="lg:hidden"
                        src="/z/pages/logo.png"
                        alt="Turnix Logo"
                        width={200}
                        height={200}
                    />
                    <div className="mt-5 space-y-2">
                        <h3 className="text-gray-800 text-2xl font-bold sm:text-3xl">Login</h3>
                    </div>
                </div>

                <form
                    onSubmit={(e) => e.preventDefault()}
                    className="space-y-5"
                >

                    <div>
                        <label className="font-medium">
                            Email
                        </label>
                        <input
                            type="email"
                            required
                            className="w-full mt-2 px-3 py-2 text-gray-500 bg-transparent outline-none border focus:border-primary-100 shadow-sm rounded-lg"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                        />
                    </div>
                    <div>
                        <label className="font-medium">
                            Password
                        </label>
                        <input
                            type="password"
                            required
                            className="w-full mt-2 px-3 py-2 text-gray-500 bg-transparent outline-none border focus:border-primary-100 shadow-sm rounded-lg"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                        />
                    </div>

                    {error && <p className="text-red-500">{error}</p>}



                    <button
                        onClick={handleSubmit}
                        disabled={loading}
                        className="w-full px-4 py-2 text-white font-medium bg-primary-700-300  rounded-lg duration-150 hover:opacity-80"
                    >
                        {loading ? "Loading..." : "Login"}
                    </button>
                </form>

                <div className="flex flex-col items-center gap-2">
                    <p className="">Need account ? <a href="/z/pages/auth/signup/open" className="font-medium text-primary-contrast-200-800">Sign up</a></p>

                    <div>
                        <a className="w-full text-xs text-white font-medium bg-secondary-700-300 duration-150 btn font-sans hover:opacity-80"
                            href="/z/pages/auth/forgot-password"

                        >
                            Forgot password
                        </a>
                    </div>



                </div>



            </div>
        </WithLoginLayout>
    </>)
}




