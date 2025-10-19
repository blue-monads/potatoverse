"use client";
import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import Image from "next/image";
import { useGApp } from "@/hooks";
import { LogOut } from "lucide-react";





export default function PortalLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {

  return (
    <>

      <div className="flex">
        <Sidebar />

        <div className="ml-14 w-full">
          {children}
        </div>

      </div>

    </>
  );
}




const Sidebar = () => {
  const [showMenu, setShowMenu] = useState(false);
  const pathname = usePathname();
  const gapp = useGApp();
  const router = useRouter();
    

  useEffect(() => {
    if (!gapp.loaded) {
      return;
    }

    if (!gapp.isAuthenticated) {
      router.push("/auth/login");
    }

  }, [gapp.isAuthenticated, gapp.loaded]);


  // console.log("@gapp", gapp);
  // console.log("isAuthenticated", gapp.isAuthenticated);
  // console.log("loaded", gapp.loaded);
  // console.log("userInfo", gapp.userInfo);

  return (
    <>
      <nav className="fixed top-0 left-0 w-14 h-full border-r border-gray-200 bg-white space-y-8">
        <div className="flex flex-col h-full">
          <div className="h-16 flex items-center justify-center mx-auto">
            <a href="/zz/pages/portal/admin" className="flex-none">
              <Image
                src="/zz/pages/logo.png"
                alt="Turnix Logo"
                width={36}
                height={36}
              />

            </a>
          </div>
          <div className="flex-1 flex flex-col h-full">
            <ul className="px-4 text-sm font-medium flex-1 flex flex-col gap-2">
              {navigation.map((item, idx) => (
                <li key={idx}>
                  <a
                    href={item.href}
                    className="relative flex items-center justify-center gap-x-2 text-gray-600 p-2 rounded-lg  hover:bg-gray-50 active:bg-gray-100 duration-150 group"
                  >
                    <div className="text-gray-500">{item.icon}</div>
                    <span className="absolute left-14 py-2 px-1.5 rounded-md whitespace-nowrap text-xs text-white bg-gray-800 hidden group-hover:inline-block group-focus:hidden duration-150">
                      {item.name}
                    </span>
                  </a>
                </li>
              ))}
            </ul>
            <div>



              <ul className="px-4 pb-4 text-sm font-medium gap-4 flex flex-col">

                {gapp.isAuthenticated && gapp.userInfo && (

                  <li>

                    <a
                      href={`/zz/pages/portal/admin/profile`}
                      className="relative flex items-center justify-center text-gray-600 rounded-lg  hover:bg-gray-50 active:bg-gray-100 duration-150 group"
                    >
                      <img src={`/zz/profileImage/11/${(gapp.userInfo.name)}`} alt="profile" className="w-8 h-8 rounded-full" />
                    </a>


                  </li>
                )}


                <li>
                  <a
                    href={`#`}
                    onClick={() => {
                      gapp.logOut();
                      router.push("/auth/login");
                    }}
                    className="relative flex items-center justify-center p-0.5 text-gray-600 rounded-lg  hover:bg-gray-50 active:bg-gray-100 duration-150 group"
                  >
                    <LogOut className="w-6 h-6" />
                  </a>
                </li>

              </ul>

            </div>
          </div>
        </div>
      </nav>
    </>
  );
};


const navigation = [
  {
    href: "/zz/pages/portal/admin",
    name: "Home",
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" aria-hidden="true" strokeWidth="1.5" viewBox="0 0 24 24" stroke="currentColor" fill="none" name="home" className="w-6 h-6">
        <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25"></path>
      </svg>
    ),
  },

  {
    href: "/zz/pages/portal/admin/spaces",
    name: "Spaces",
    icon: (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        strokeWidth={1.5}
        stroke="currentColor"
        className="w-6 h-6"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M6.429 9.75L2.25 12l4.179 2.25m0-4.5l5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0l4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0l-5.571 3-5.571-3"
        />
      </svg>
    ),
  },
  {
    href: "/zz/pages/portal/admin/store",
    name: "Store",
    icon: (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        strokeWidth={1.5}
        stroke="currentColor"
        className="w-6 h-6"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M14.25 6.087c0-.355.186-.676.401-.959.221-.29.349-.634.349-1.003 0-1.036-1.007-1.875-2.25-1.875s-2.25.84-2.25 1.875c0 .369.128.713.349 1.003.215.283.401.604.401.959v0a.64.64 0 01-.657.643 48.39 48.39 0 01-4.163-.3c.186 1.613.293 3.25.315 4.907a.656.656 0 01-.658.663v0c-.355 0-.676-.186-.959-.401a1.647 1.647 0 00-1.003-.349c-1.036 0-1.875 1.007-1.875 2.25s.84 2.25 1.875 2.25c.369 0 .713-.128 1.003-.349.283-.215.604-.401.959-.401v0c.31 0 .555.26.532.57a48.039 48.039 0 01-.642 5.056c1.518.19 3.058.309 4.616.354a.64.64 0 00.657-.643v0c0-.355-.186-.676-.401-.959a1.647 1.647 0 01-.349-1.003c0-1.035 1.008-1.875 2.25-1.875 1.243 0 2.25.84 2.25 1.875 0 .369-.128.713-.349 1.003-.215.283-.4.604-.4.959v0c0 .333.277.599.61.58a48.1 48.1 0 005.427-.63 48.05 48.05 0 00.582-4.717.532.532 0 00-.533-.57v0c-.355 0-.676.186-.959.401-.29.221-.634.349-1.003.349-1.035 0-1.875-1.007-1.875-2.25s.84-2.25 1.875-2.25c.37 0 .713.128 1.003.349.283.215.604.401.96.401v0a.656.656 0 00.658-.663 48.422 48.422 0 00-.37-5.36c-1.886.342-3.81.574-5.766.689a.578.578 0 01-.61-.58v0z"
        />
      </svg>
    ),
  },

  {
    href: "/zz/pages/portal/admin/users",
    name: "Users",
    icon: (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        strokeWidth={1.5}
        stroke="currentColor"
        className="w-6 h-6"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M16.5 7.5a3 3 0 11-6 0 3 3 0 016 0zm-6 3a6.75 6.75 0 00-6.75 6.75v1.5A2.25 2.25 0 006.75 21h10.5a2.25 2.25 0 002.25-2.25v-1.5A6.75 6.75 0 0010.5 10.5zM12 12a4.5 4.5 0 100-9 4.5 4.5 0 000 9z"
        />

      </svg>
    ),

  },


  {
    href: "#",
    name: "Settings",
    icon: (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        strokeWidth={1.5}
        stroke="currentColor"
        className="w-5 h-5"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z"
        />
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
        />
      </svg>
    ),
  },

];
