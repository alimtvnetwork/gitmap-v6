import { useState, useCallback } from "react";
import { Copy, Check, Terminal } from "lucide-react";

interface InstallTab {
  label: string;
  command: string;
}

interface InstallBlockProps {
  command?: string;
  tabs?: InstallTab[];
}

const CopyLine = ({ command }: { command: string }) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(command);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [command]);

  return (
    <div
      onClick={handleCopy}
      className="flex items-start gap-3 px-5 py-3 rounded-lg bg-terminal border border-border cursor-pointer hover:border-primary/40 transition-colors group"
    >
      <Terminal className="h-4 w-4 text-primary mt-0.5 shrink-0" />
      <code className="font-mono text-sm text-terminal-foreground break-all leading-relaxed flex-1">
        {command}
      </code>
      <span className="text-muted-foreground group-hover:text-foreground transition-colors shrink-0 mt-0.5">
        {copied ? <Check className="h-4 w-4 text-primary" /> : <Copy className="h-4 w-4" />}
      </span>
    </div>
  );
};

const InstallBlock = ({ command, tabs }: InstallBlockProps) => {
  const [active, setActive] = useState(0);

  if (tabs && tabs.length > 0) {
    return (
      <div className="space-y-2">
        <div className="flex gap-1 justify-center">
          {tabs.map((tab, i) => (
            <button
              key={tab.label}
              onClick={() => setActive(i)}
              className={`px-3 py-1 rounded-md text-xs font-mono transition-colors ${
                i === active
                  ? "bg-primary text-primary-foreground"
                  : "bg-muted text-muted-foreground hover:text-foreground"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
        <CopyLine command={tabs[active].command} />
      </div>
    );
  }

  if (command) {
    return <CopyLine command={command} />;
  }

  return null;
};

export default InstallBlock;
