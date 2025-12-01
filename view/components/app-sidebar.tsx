"use client";

import * as React from "react";

import { NavMain } from "@/components/nav-main";
import { NavSecondary } from "@/components/nav-secondary";
import { NavUser } from "@/components/nav-user";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import {
  Bitcoin,
  HelpCircle,
  Search,
  Send,
  Settings,
  WalletIcon,
} from "lucide-react";
import { useAuthStore } from "@/store/auth-store";

const data = {
  user: {
    name: "shadcn",
    email: "m@example.com",
    avatar: "/avatars/shadcn.jpg",
  },
  navMain: [
    {
      title: "Wallet",
      url: "/dashboard",
      icon: <WalletIcon />,
    },
    {
      title: "Send",
      url: "/dashboard/send",
      icon: <Send />,
    },
  ],
  navSecondary: [
    {
      title: "Settings",
      url: "#",
      icon: <Settings />,
    },
    {
      title: "Get Help",
      url: "#",
      icon: <HelpCircle />,
    },
    {
      title: "Search",
      url: "#",
      icon: <Search />,
    },
  ],
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const userAuth = useAuthStore((state) => state.user);
  const [user, setUser] = React.useState(userAuth);

  React.useEffect(() => {
    setUser(userAuth);
  }, [userAuth]);

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:!p-1.5"
            >
              <a href="#">
                <Bitcoin />
                <span className="text-base font-semibold">Yute Blockchain</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={user} />
      </SidebarFooter>
    </Sidebar>
  );
}
