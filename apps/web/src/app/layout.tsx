import type { Metadata } from "next";
import { Fredoka } from "next/font/google";
import { ReactQueryProvider } from "@/lib/query-client";
import "./globals.css";

const fredoka = Fredoka({
  variable: "--font-fredoka",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: "ShortKing - Shorten Links Like a King",
  description: "A URL shortener with custom aliases, analytics, and QR codes.",
  icons: {
    icon: "/logo-king.png",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={`${fredoka.variable} h-full antialiased`}>
      <body className="min-h-full flex flex-col">
        <ReactQueryProvider>{children}</ReactQueryProvider>
      </body>
    </html>
  );
}
