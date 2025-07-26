"use client";
import { useState } from "react";
import { usePathname } from "next/navigation";
import Image from "next/image";


export default function PortalLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const [showMenu, setShowMenu] = useState(false);
  const pathname = usePathname();

  return (
    <>
      <nav className="bg-white border border-gray-200 dark:border-gray-700 px-2 sm:px-4 py-2.5 rounded dark:bg-gray-800 shadow">
        <div className="container flex flex-wrap justify-between items-center mx-auto">
          <a href="/pages/portal" className="flex items-center gap-1">

            <Image
              className="dark:invert p-1 border "
              src="/logo.png"
              alt="Turnix Logo"
              width={42}
              height={42}
              priority
            />

            <span className="self-center text-xl font-semibold whitespace-nowrap dark:text-white">
              Turnix Portal
            </span>
          </a>

          <div className="flex gap-2 items-center ">
            <div className="flex-none h-full text-center flex items-center justify-center">

              <div className="flex items-center justify-center mr-2">
                <button
                  className="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 focus:outline-none"

                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="size-6">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0" />
                  </svg>

                </button>
              </div>

              <div className="flex space-x-3 items-center px-1">
                <div className="flex-none flex justify-center ">
                  <div className="w-8 h-8 flex ">
                    <img src="https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcShta_GXR2xdnsxSzj_GTcJHcNykjVKrCBrZ9qouUl0usuJWG2Rpr_PbTDu3sA9auNUH64&usqp=CAU" alt="profile" className="shadow rounded-full object-cover" />
                  </div>
                </div>
                <div className="hidden md:block text-sm md:text-md text-black dark:text-white">John Doe</div>
              </div>

              {/* Notification icon */}
              
            </div>

          </div>
        </div>
      </nav>

      <div className="container mx-auto p-4">
        {children}
      </div>
    </>
  );
}