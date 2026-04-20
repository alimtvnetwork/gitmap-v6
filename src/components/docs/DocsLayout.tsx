import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { DocsSidebar } from "@/components/docs/DocsSidebar";
import CommandPalette from "@/components/docs/CommandPalette";
import { VERSION } from "@/constants/index";

interface DocsLayoutProps {
  children: React.ReactNode;
}

const DocsLayout = ({ children }: DocsLayoutProps) => {
  return (
    <SidebarProvider>
      <div className="min-h-screen flex w-full">
        <DocsSidebar />
        <div className="flex-1 flex flex-col min-w-0">
          <header className="h-12 flex items-center border-b border-border sticky top-0 bg-background/80 backdrop-blur-sm z-10">
            <SidebarTrigger className="ml-3" />
<span className="ml-3 text-sm font-mono text-muted-foreground">gitmap documentation</span>
            <span className="ml-2 px-2 py-0.5 text-xs font-mono bg-primary/10 text-primary rounded">{VERSION}</span>
            <div className="ml-auto mr-3">
              <CommandPalette />
            </div>
          </header>
          <main className="flex-1 overflow-auto">
            <div className="max-w-4xl mx-auto px-6 py-8">
              {children}
            </div>
          </main>
        </div>
      </div>
    </SidebarProvider>
  );
};

export default DocsLayout;
