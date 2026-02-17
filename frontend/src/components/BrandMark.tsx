import clsx from "clsx";

export function BrandMark({
  label,
  className,
  size = "md",
}: {
  label: string;
  className?: string;
  size?: "sm" | "md" | "lg";
}) {
  const sizeClass =
    size === "sm" ? "h-8 w-8 text-xs" : size === "lg" ? "h-10 w-10 text-sm" : "h-9 w-9 text-xs";

  return (
    <span
      role="img"
      aria-label={label}
      className={clsx(
        "inline-flex items-center justify-center rounded-full border border-neutral-200 shadow-sm",
        "brand-gradient text-white",
        sizeClass,
        className,
      )}
    >
      <span aria-hidden="true" className="font-semibold">
        {label.trim().charAt(0).toUpperCase() || "V"}
      </span>
    </span>
  );
}

