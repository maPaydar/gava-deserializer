import java.io.*;
import java.util.*;

public class Test implements Serializable {
    private String a;
    private int b;
    private byte[] c;

    public Test(String a, int b, byte[] c) {
        this.a = a;
        this.b = b;
        this.c = c;
    }

    private static final char[] HEX_ARRAY = "0123456789ABCDEF".toCharArray();

    public static String bytesToHex(byte[] bytes) {
        char[] hexChars = new char[bytes.length * 2];
        for (int j = 0; j < bytes.length; j++) {
            int v = bytes[j] & 0xFF;
            hexChars[j * 2] = HEX_ARRAY[v >>> 4];
            hexChars[j * 2 + 1] = HEX_ARRAY[v & 0x0F];
        }
        return new String(hexChars);
    }

    public static void main(String[] args) throws Exception {
        Test test = new Test("aa", 1, new byte[]{1,2,3,4});
        FileOutputStream fileOutputStream
              = new FileOutputStream("test.txt");
        ObjectOutputStream objectOutputStream
          = new ObjectOutputStream(fileOutputStream);
        objectOutputStream.writeObject(test);
        objectOutputStream.flush();
        objectOutputStream.close();
        fileOutputStream.flush();
        fileOutputStream.close();
        Scanner scanner = new Scanner(new File("test.txt"));
        String s = scanner.nextLine();
        String hex = bytesToHex(s.getBytes());
        System.out.println(hex);
    }
}