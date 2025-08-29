"use client"
import Image from "next/image";
import WithLoginLayout from "./WithLoginLayout";
import { useState } from "react";


export default function Page() {

    const [email, setEmail] = useState<string>("");
    const [password, setPassword] = useState<string>("");

    const handleSubmit = (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();
        console.log(email, password);
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
                    <button
                        onClick={handleSubmit}
                        className="w-full px-4 py-2 text-white font-medium bg-primary-700-300  rounded-lg duration-150"
                    >
                        Create account
                    </button>
                </form>

                <div className="flex flex-col items-center gap-2">
                    <p className="">Need account ? <a href="/z/pages/auth/signup/open" className="font-medium text-primary-contrast-200-800">Sign up</a></p>

                    <div>
                        <button className="w-full px-2 py-1 text-xs text-white font-medium bg-secondary-700-300 rounded duration-150 btn">
                            Forgot password
                        </button>
                    </div>



                </div>



            </div>
        </WithLoginLayout>
    </>)
}




