"use client"
import Image from "next/image";
import { useState } from "react";
export default function Page() {
    return (<>
        <AuthPage />
    </>)
}


const AuthPage = () => {
    return (
        <main className="w-full flex">
            <div className="relative flex-1 hidden items-center justify-center h-screen bg-gray-900 lg:flex">
                <div className="relative z-10 w-full max-w-md">
                    <Image
                        src="/z/pages/logo.png"
                        alt="Turnix Logo"
                        width={200}
                        height={200}
                    />
                    <div className=" mt-16 space-y-3">
                        <h3 className="text-white text-3xl font-bold">Turnix</h3>
                        <p className="text-gray-300">
                            Platfrom for your apps.
                        </p>

                    </div>
                </div>
                <div
                    className="absolute inset-0 my-auto h-[500px]"
                    style={{
                        background: "linear-gradient(152.92deg, rgba(192, 132, 252, 0.2) 4.54%, rgba(232, 121, 249, 0.26) 34.2%, rgba(192, 132, 252, 0.1) 77.55%)", filter: "blur(118px)"
                    }}
                >

                </div>
            </div>
            <div className="flex-1 flex items-center justify-center h-screen">
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
                                className="w-full mt-2 px-3 py-2 text-gray-500 bg-transparent outline-none border focus:border-indigo-600 shadow-sm rounded-lg"
                            />
                        </div>
                        <div>
                            <label className="font-medium">
                                Password
                            </label>
                            <input
                                type="password"
                                required
                                className="w-full mt-2 px-3 py-2 text-gray-500 bg-transparent outline-none border focus:border-indigo-600 shadow-sm rounded-lg"
                            />
                        </div>
                        <button
                            className="w-full px-4 py-2 text-white font-medium bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-600 rounded-lg duration-150"
                        >
                            Create account
                        </button>
                    </form>

                    <p className="">Need account ? <a href="javascript:void(0)" className="font-medium text-indigo-600 hover:text-indigo-500">Sign up</a></p>
                </div>
            </div>
        </main>
    )
}