"use client";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldSeparator,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import Link from "next/link";
import { Textarea } from "./ui/textarea";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import { useLogin } from "@/hooks/use-login";

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const {
    mnemonic,
    password,
    isPasswordVisible,
    file,
    isLoading,
    setMnemonic,
    setPassword,
    setFile,
    handleLogin,
    togglePasswordVisibility,
    username,
    setUsername,
  } = useLogin();

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card className="overflow-hidden p-0">
        <CardContent className="grid p-0 md:grid-cols-2">
          <div className="p-6 md:p-8">
            <h1 className="text-center text-2xl font-semibold">Login</h1>
            <p className="text-center mb-6">
              Please enter your credentials to access your account.
            </p>

            <form onSubmit={handleLogin}>
              <FieldGroup className="mb-6 gap-2">
                <Input
                  type="text"
                  placeholder="Enter your username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                />
              </FieldGroup>

              {file === null && (
                <FieldGroup className="mb-6 gap-2">
                  <Textarea
                    rows={4}
                    placeholder="Enter your mnemonic phrase"
                    value={mnemonic}
                    onChange={(e) => setMnemonic(e.target.value)}
                  />
                </FieldGroup>
              )}

              <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card my-6">
                Or use your wallet file
              </FieldSeparator>

              <FieldGroup className="mb-6 gap-2">
                <Input
                  type="file"
                  accept=".json"
                  onChange={(e) => {
                    if (e.target.files && e.target.files.length > 0) {
                      setFile(e.target.files[0]);
                    }
                  }}
                />
              </FieldGroup>

              {file && (
                <FieldGroup className="mb-6 gap-2">
                  <div className="relative">
                    <Input
                      type={isPasswordVisible ? "text" : "password"}
                      placeholder="Password"
                      className="pr-9"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={togglePasswordVisibility}
                      className="text-muted-foreground focus-visible:ring-ring/50 absolute inset-y-0 right-0 rounded-l-none hover:bg-transparent"
                    >
                      {isPasswordVisible ? <EyeOffIcon /> : <EyeIcon />}
                      <span className="sr-only">
                        {isPasswordVisible ? "Hide password" : "Show password"}
                      </span>
                    </Button>
                  </div>
                </FieldGroup>
              )}

              <FieldGroup>
                <Field>
                  <Button type="submit" disabled={isLoading}>
                    {isLoading ? "Logging in..." : "Login"}
                  </Button>
                </Field>
                <FieldDescription className="text-center">
                  Don&apos;t have an wallet?{" "}
                  <Link href="/signup">Create Wallet</Link>
                </FieldDescription>
              </FieldGroup>
            </form>
          </div>
          <div className="bg-muted relative hidden md:block">
            <img
              src="https://res.cloudinary.com/dwg1vtwlc/image/upload/v1764055412/409111383_01f7bf28-3367-47b9-a49e-d4fd9814f722_i1mb7p.jpg"
              alt="Image"
              className="absolute inset-0 h-full w-full object-cover dark:brightness-[0.2] dark:grayscale"
            />
          </div>
        </CardContent>
      </Card>
      <FieldDescription className="px-6 text-center">
        By clicking continue, you agree to our <a href="#">Terms of Service</a>{" "}
        and <a href="#">Privacy Policy</a>.
      </FieldDescription>
    </div>
  );
}
