/* eslint-disable @typescript-eslint/no-explicit-any */
"use client";

import { useState } from "react";
import {
  useGetAdmins,
  useCreateAdmin,
  useDeleteAdmin,
} from "@/hooks/use-admin";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Loader2, Plus, Trash2 } from "lucide-react";

export default function AdminsPage() {
  const [newAdminUserId, setNewAdminUserId] = useState("");
  const [newAdminRole, setNewAdminRole] = useState<
    "admin" | "moderator" | "support"
  >("moderator");
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);

  const { data: admins, isLoading } = useGetAdmins();
  const { mutate: createAdmin, isPending: isCreating } = useCreateAdmin();
  const { mutate: deleteAdmin, isPending: isDeleting } = useDeleteAdmin();

  const handleCreateAdmin = () => {
    if (!newAdminUserId) return;

    createAdmin(
      {
        user_id: parseInt(newAdminUserId),
        role: newAdminRole,
        permissions:
          newAdminRole === "admin"
            ? ["*"]
            : newAdminRole === "moderator"
              ? ["read_dashboard", "view_activity_logs", "read_users"]
              : ["read_users", "view_activity_logs"],
      },
      {
        onSuccess: () => {
          setNewAdminUserId("");
          setNewAdminRole("moderator");
          setIsCreateDialogOpen(false);
        },
      },
    );
  };

  const statusColors: Record<string, string> = {
    active: "bg-green-100 text-green-800",
    inactive: "bg-gray-100 text-gray-800",
    suspended: "bg-red-100 text-red-800",
  };

  const roleColors: Record<string, string> = {
    admin: "bg-purple-100 text-purple-800",
    moderator: "bg-blue-100 text-blue-800",
    support: "bg-yellow-100 text-yellow-800",
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            Admin Management
          </h1>
          <p className="text-gray-500 mt-2">Kelola akun admin dan role</p>
        </div>

        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Add Admin
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Admin</DialogTitle>
              <DialogDescription>
                Promosikan user menjadi admin dengan role dan permissions yang
                sesuai
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="user_id">User ID</Label>
                <Input
                  id="user_id"
                  type="number"
                  placeholder="Masukkan User ID"
                  value={newAdminUserId}
                  onChange={(e) => setNewAdminUserId(e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="role">Role</Label>
                <Select
                  value={newAdminRole}
                  onValueChange={(v: any) => setNewAdminRole(v)}
                >
                  <SelectTrigger id="role">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="admin">Admin</SelectItem>
                    <SelectItem value="moderator">Moderator</SelectItem>
                    <SelectItem value="support">Support</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsCreateDialogOpen(false)}
              >
                Cancel
              </Button>
              <Button
                onClick={handleCreateAdmin}
                disabled={isCreating || !newAdminUserId}
              >
                {isCreating ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Creating...
                  </>
                ) : (
                  "Create Admin"
                )}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Admin List</CardTitle>
          <CardDescription>Total: {admins?.length || 0} admins</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {[...Array(5)].map((_, i) => (
                <Skeleton key={i} className="h-12" />
              ))}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>Username</TableHead>
                    <TableHead>Address</TableHead>
                    <TableHead>Role</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Last Login</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {admins && admins.length > 0 ? (
                    admins.map((admin) => (
                      <TableRow key={admin.id}>
                        <TableCell className="font-medium">
                          {admin.id}
                        </TableCell>
                        <TableCell>{admin.username}</TableCell>
                        <TableCell className="font-mono text-sm">
                          {admin.address.slice(0, 10)}...
                        </TableCell>
                        <TableCell>
                          <Badge className={roleColors[admin.role]}>
                            {admin.role}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className={statusColors[admin.status]}>
                            {admin.status}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {admin.last_login_at
                            ? new Date(admin.last_login_at).toLocaleDateString(
                                "id-ID",
                              )
                            : "-"}
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => deleteAdmin(admin.id)}
                            disabled={isDeleting}
                          >
                            {isDeleting ? (
                              <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                              <Trash2 className="h-4 w-4" />
                            )}
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell
                        colSpan={7}
                        className="text-center py-8 text-gray-500"
                      >
                        No admins found
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
