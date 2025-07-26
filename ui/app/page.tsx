import Image from "next/image";

export default function Home() {
  return (
    <div className="font-sans grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20">
      <div className="mt-10 py-10">
        <Image
          src="/logo.png"
          alt="Turnix Logo"
          width={200}
          height={200}
        />

      </div>

    </div>
  );
}
