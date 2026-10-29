@IsTest
private class GreeterTest {
    @IsTest
    static void greetsByName() {
        System.assertEquals('Hello, Ada', Greeter.greet('Ada'));
    }
}
