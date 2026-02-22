"use client";

import Link from "next/link";
import type { ReactNode } from "react";
import clsx from "clsx";

type Button3DProps = {
  children: ReactNode;
  variant?: "primary" | "secondary";
  href?: string;
  onClick?: () => void;
  type?: "button" | "submit";
  className?: string;
};

export function Button3D({
  children,
  variant = "secondary",
  href,
  onClick,
  type = "button",
  className,
}: Button3DProps) {
  const outerClass = clsx(
    "btn-3d-outer inline-block cursor-pointer rounded-[0.3rem] border-2 border-white p-[0.1rem] text-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-white/60 focus-visible:ring-offset-2 focus-visible:ring-offset-[#2b2929]",
    variant === "primary" && "bg-[#212121]",
    variant === "secondary" && "bg-[#212121]",
    className,
  );

  const innerClass = clsx(
    "btn-3d-inner h-12 min-w-[8rem] px-6 font-semibold text-white sm:min-w-[10rem] bg-[#212121] border-white",
  );

  const content = <span className={innerClass}>{children}</span>;

  if (href) {
    return (
      <Link href={href} className={outerClass}>
        {content}
      </Link>
    );
  }

  return (
    <button type={type} onClick={onClick} className={outerClass}>
      {content}
    </button>
  );
}
