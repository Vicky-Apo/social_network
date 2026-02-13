"use client";

import {
  motion,
  type HTMLMotionProps,
  type Variants,
  useReducedMotion,
} from "framer-motion";
import type { PropsWithChildren, ReactNode } from "react";

export const fadeUp: Variants = {
  hidden: { opacity: 0, y: 16 },
  show: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.55, ease: [0.22, 1, 0.36, 1] },
  },
};

export const staggerContainer: Variants = {
  hidden: {},
  show: {
    transition: {
      staggerChildren: 0.12,
      delayChildren: 0.05,
    },
  },
};

export const viewportOnce = {
  once: true,
  amount: 0.2,
};

type RevealProps = PropsWithChildren<HTMLMotionProps<"div">> & {
  delay?: number;
};

export function MotionReveal({
  children,
  delay = 0,
  variants,
  transition,
  ...props
}: RevealProps) {
  const reducedMotion = useReducedMotion();

  if (reducedMotion) {
    return <motion.div {...props}>{children}</motion.div>;
  }

  return (
    <motion.div
      initial="hidden"
      whileInView="show"
      viewport={viewportOnce}
      variants={variants ?? fadeUp}
      transition={transition ?? { delay }}
      {...props}
    >
      {children}
    </motion.div>
  );
}

export function MotionFloat({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  const reducedMotion = useReducedMotion();

  if (reducedMotion) {
    return <div className={className}>{children}</div>;
  }

  return (
    <motion.div
      className={className}
      animate={{ y: [0, -8, 0] }}
      transition={{ duration: 7.5, ease: "easeInOut", repeat: Infinity }}
    >
      {children}
    </motion.div>
  );
}
