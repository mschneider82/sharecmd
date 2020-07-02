# This file was generated by GoReleaser. DO NOT EDIT.
class Sharecmd < Formula
  desc "Share your files using Cloudproviders with just one command"
  homepage "https://github.com/mschneider82/sharecmd"
  version "0.0.106"
  bottle :unneeded

  if OS.mac?
    url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.106/sharecmd_0.0.106_Darwin_x86_64.tar.gz"
    sha256 "64b6e62011d465f21a161f0413da46ded8c9157ea6f8d2a820a61bd92f9349eb"
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.106/sharecmd_0.0.106_Linux_x86_64.tar.gz"
      sha256 "44add7be72f7d9bca7ace743f2ed707b883326aee854fda87dc984fbbd7ddfc3"
    end
    if Hardware::CPU.arm?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.106/sharecmd_0.0.106_Linux_arm64.tar.gz"
        sha256 "3ec3e866245b54a9cca605beb5056334752aa1ad457d1855e5cf3f88bc71ee2f"
      else
        url "https://github.com/mschneider82/sharecmd/releases/download/v0.0.106/sharecmd_0.0.106_Linux_armv6.tar.gz"
        sha256 "451f91b087a01ce60d282448bc740b81affd86c3637e96ff2872cc8017611589"
      end
    end
  end

  def install
    bin.install "share"
  end

  test do
    system "#{bin}/share --help"
  end
end
