import { AuthGuard } from "@/components/AuthGuard";
import { Sidebar } from "@/components/Sidebar";
import { UserHeader } from "@/components/UserHeader";

export default function BranchLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthGuard allowedRoles={["BRANCH_MANAGER", "ADMIN", "SUPER_ADMIN"]}>
      <div className="flex h-screen bg-gray-50">
        <Sidebar />
        <div className="flex-1 flex flex-col overflow-hidden">
          <UserHeader />
          <main className="flex-1 overflow-auto">{children}</main>
        </div>
      </div>
    </AuthGuard>
  );
}
