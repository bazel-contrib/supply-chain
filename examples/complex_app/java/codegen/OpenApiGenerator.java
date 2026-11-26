package demo.codegen;

import java.nio.file.Files;
import java.nio.file.Path;

public final class OpenApiGenerator {
    public static void main(String[] args) throws Exception {
        String spec = Files.readString(Path.of(args[0])).trim();
        String goSource = "package main\n\nconst openAPISpec = `" + spec + "`\n";
        Files.writeString(Path.of(args[1]), goSource);
    }
}
