import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Turnix",
  description: "Ultimate thing maker",
};


const staticGradients = [
  'from-pink-500 to-orange-500',
  'from-blue-500 to-purple-600',
  'from-gray-600 to-blue-800',
  'from-red-500 to-pink-600',
  'from-purple-600 to-indigo-600',
  'from-purple-500 to-pink-500',
  'from-teal-500 to-blue-600',
  'from-green-500 to-blue-600'
]

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


        {children}
      </body>
    </html>
  );
}
