import clsx from "clsx";
import Image from "next/image";

const sizeMap = {
  sm: { container: "h-8", image: 32 },
  md: { container: "h-9", image: 36 },
  lg: { container: "h-10", image: 40 },
} as const;

export function BrandMark({
  label,
  className,
  size = "md",
  logoSrc,
}: {
  label: string;
  className?: string;
  size?: "sm" | "md" | "lg";
  logoSrc?: string;
}) {
  const { container, image } = sizeMap[size];

  if (logoSrc) {
    return (
      <span
        role="img"
        aria-label={label}
        className={clsx("relative inline-flex shrink-0", container, className)}
      >
        <Image
          src={logoSrc}
          alt={label}
          width={image * 2.4}
          height={image}
          className="h-full w-auto object-contain object-left"
          priority
        />
      </span>
    );
  }

  return (
    <span
      role="img"
      aria-label={label}
      className={clsx(
        "inline-flex items-center justify-center rounded-sm border border-neutral-200 shadow-sm",
        "brand-gradient text-white",
        size === "sm" ? "h-8 w-8 text-xs" : size === "lg" ? "h-10 w-10 text-sm" : "h-9 w-9 text-xs",
        className,
      )}
    >
      <span aria-hidden="true" className="font-semibold">
        {label.trim().charAt(0).toUpperCase() || "V"}
      </span>
    </span>
  );
}

