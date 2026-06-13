import { ButtonHTMLAttributes, forwardRef } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "danger";
  size?: "sm" | "md" | "lg";
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = "primary", size = "md", className = "", ...props }, ref) => {
    const base = "inline-flex items-center justify-center rounded-lg font-medium transition focus:outline-none focus:ring-2";
    const variants = {
      primary: "bg-mesob-blue text-white hover:opacity-90 focus:ring-mesob-blue",
      secondary: "bg-white text-gray-700 border border-gray-300 hover:bg-gray-50 focus:ring-gray-300",
      danger: "bg-red-600 text-white hover:bg-red-700 focus:ring-red-500",
    };
    const sizes = {
      sm: "px-3 py-1.5 text-sm",
      md: "px-4 py-2 text-sm",
      lg: "px-6 py-3 text-base",
    };
    return (
      <button
        ref={ref}
        className={`${base} ${variants[variant]} ${sizes[size]} ${className}`}
        {...props}
      />
    );
  }
);
Button.displayName = "Button";
