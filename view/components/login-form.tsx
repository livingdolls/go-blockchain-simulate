"use client";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { ethers, hashMessage, verifyMessage } from "ethers";
import { useState } from "react";
import { WalletFromMnemonic } from "@/lib/crypto";
import { api } from "@/lib/axios";
import Link from "next/link";

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const [mnemonic, setMnemonic] = useState(
    "spatial media crunch crop clump candy rotate hollow amount tissue total scene"
  );

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!mnemonic) {
      alert("Please enter your mnemonic.");
      return;
    }

    const wallet = WalletFromMnemonic(mnemonic);
    const addr = wallet.address.toLowerCase();

    console.log("Logging in with address:", addr);

    const ch = await api.post(`/challenge/${addr}`);
    const nonce = ch.data.challenge;
    if (!nonce) {
      alert("Failed to get challenge from server.");
      return;
    }

    // sign cannonical message
    const message = `Login to YuteBlockchain nonce:${nonce}`;

    const hashedMessage = hashMessage(message);
    console.log("Message to sign:", message);
    console.log("Hashed message:", hashedMessage);

    const signature = await wallet.signMessage(message);
    console.log("Signature:", signature);
    console.log("Last byte of signature (v):", signature.slice(-2));

    // verification of signed message
    const recovered = verifyMessage(message, signature);

    console.log("Recovered address:", recovered);
    console.log("Derived address", addr);
    console.log("Match", recovered.toLowerCase() === addr);

    const payload = {
      address: addr,
      signature,
      nonce,
    };

    // send signature to server for verification
    const res = await api.post(`/challenge/verify`, payload);

    console.log("Login response:", res);

    if (res.data.valid) {
      window.location.href = "/dashboard";
    }
  };

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card className="overflow-hidden p-0">
        <CardContent className="grid p-0 md:grid-cols-2">
          <form className="p-6 md:p-8" onSubmit={handleLogin}>
            <FieldGroup>
              <div className="flex flex-col items-center gap-2 text-center">
                <h1 className="text-2xl font-bold">Welcome back</h1>
                <p className="text-muted-foreground text-balance">
                  Login to your Wallet Account
                </p>
              </div>
              <Field>
                <Button type="submit">Login</Button>
              </Field>
              <FieldDescription className="text-center">
                Don&apos;t have an wallet?{" "}
                <Link href="/signup">Create Wallet</Link>
              </FieldDescription>
            </FieldGroup>
          </form>
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
