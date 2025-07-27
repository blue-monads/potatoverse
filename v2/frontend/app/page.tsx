import Image from "next/image";

export default function Home() {
  return (
    <div className="font-sans grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20">

      <ul className="list-disc list-inside">
        <li>
          <a className="hover:underline" href="/z/pages/admin">Admin</a>
        </li>
        <li>
          <a className="hover:underline" href="/z/pages/portal">Portal</a>
        </li>
        <li>
          <a className="hover:underline" href="/z/pages/auth">Login</a>
        </li>

        <li>
          <a className="hover:underline" href="/z/pages/theme">Theme</a>
        </li>


      </ul>

    </div>
  );
}
