"use client"
import "./globals.css";
import { GAppStateContext } from "@/hooks/contexts/GAppStateContext";
import GModalWrapper from "@/hooks/modal/GModalWrapper";
import { staticGradients } from "./utils";



export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      data-theme="cerberus"
    >
      <body
        className={`antialiased`}
      >

        <div className="hidden">
          {staticGradients.map((gradient, index) => (
            <div key={index} className={`bg-gradient-to-br ${gradient} h-0 w-0 absolute`} />
          ))}
        </div>

        <GAppStateContext>
          {children}
          <GModalWrapper />
        </GAppStateContext>
      </body>
    </html>
  );
}
