using RoslynHelper.JsonRpc;
using RoslynHelper.Workspace;

namespace RoslynHelper;

public static class Program
{
    public static async Task<int> Main(string[] args)
    {
        // UTF-8 without BOM on both ends of the pipe.
        Console.InputEncoding  = new System.Text.UTF8Encoding(encoderShouldEmitUTF8Identifier: false);
        Console.OutputEncoding = new System.Text.UTF8Encoding(encoderShouldEmitUTF8Identifier: false);

        using var workspace = new WorkspaceManager();
        var dispatcher = new Dispatcher(workspace);

        await dispatcher.RunAsync(Console.In, Console.Out);
        return 0;
    }
}
