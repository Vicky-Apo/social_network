import type { PropsWithChildren } from "react";
import clsx from "clsx";

type SectionProps = PropsWithChildren<{
  id?: string;
  className?: string;
  containerClassName?: string;
}>;

export function Section({
  id,
  className,
  containerClassName,
  children,
}: SectionProps) {
  return (
    <section
      id={id}
      className={clsx("scroll-mt-28 py-16 md:py-20", className)}
    >
      <div className={clsx("mx-auto w-full max-w-6xl px-4 sm:px-6", containerClassName)}>
        {children}
      </div>
    </section>
  );
}
