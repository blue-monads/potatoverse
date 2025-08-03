import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Turnix",
  description: "Ultimate thing maker",
};

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
        {children}
      </body>
    </html>
  );
}
