"use client";

type Props = {
  src?: string | null;
  name?: string | null;
  size?: number;
  className?: string;
  textClassName?: string;
};

function initials(name?: string | null) {
  if (!name) return "U";
  const parts = name.trim().split(/\s+/);
  const first = parts[0]?.[0] ?? "";
  const second = parts.length > 1 ? parts[parts.length - 1]?.[0] ?? "" : "";
  const out = `${first}${second}`.toUpperCase();
  return out || "U";
}

export default function Avatar({
  src,
  name,
  size = 40,
  className,
  textClassName,
}: Props) {
  const dimension = { width: size, height: size };
  const wrapperClass = ["rounded-full", "border", "border-neutral-200", "bg-white", "overflow-hidden", className]
    .filter(Boolean)
    .join(" ");
  const fallbackClass = [
    "inline-flex",
    "items-center",
    "justify-center",
    "rounded-full",
    "bg-neutral-900",
    "text-white",
    className,
  ]
    .filter(Boolean)
    .join(" ");
  const textClass = ["text-xs", "font-semibold", textClassName].filter(Boolean).join(" ");

  if (src) {
    return (
      <div className={wrapperClass} style={dimension}>
        <img src={src} alt={name ?? "Avatar"} className="h-full w-full object-cover" />
      </div>
    );
  }

  return (
    <div className={fallbackClass} style={dimension}>
      <span className={textClass}>{initials(name)}</span>
    </div>
  );
}
