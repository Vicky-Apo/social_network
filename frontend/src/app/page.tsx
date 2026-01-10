import Link from "next/link";
import Image from "next/image";

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-white text-neutral-900 flex flex-col gap-16">

      {/* Nav */}
      <nav className="border-b border-neutral-200 mt-12">
        <div className="max-w-7xl mx-auto pl-4 md:pl-10 pr-6 py-6">
          <div className="flex items-center justify-between">
            <Link href="/" className="flex items-center gap-3">
              <span className="text-base font-semibold tracking-tight">
                ConNextioN
              </span>
            </Link>

            <div className="flex items-center gap-8">
              <Link
                href="/login"
                className="text-sm text-neutral-600 hover:text-neutral-900 transition"
              >
                Sign In
              </Link>
              <Link
                href="/register"
                className="text-sm font-medium text-neutral-900 underline underline-offset-8 decoration-2 decoration-neutral-900/30 hover:decoration-neutral-900 transition"
              >
                Sign Up
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero */}
      <section >
        <div className="max-w-7xl mx-auto pl-4 md:pl-10 pr-6 pt-20 pb-32 md:pt-28 md:pb-36">
          <div className="grid lg:grid-cols-2 gap-12 lg:gap-16 items-start">
            {/* Left */}
            <div className="max-w-2xl">
              <h1 className="text-5xl md:text-6xl font-semibold leading-[1.05] tracking-tight">
                Next Gen
                <br />
                <span className="font-black">Connection</span>
                <br />
                without limitations
              </h1>
              <div className="mt-10 md:ml-24">
                <p className="text-lg md:text-xl text-neutral-600 leading-relaxed max-w-md">
                  Connect with those who matter
                  <br />
                  Explore communities
                  <br />
                  Interact with like-minded people
                  <br />
                  Share ideas, collaborate, and grow together
                </p>
              </div>
            </div>

            {/* Right */}
            <div className="flex justify-center -mt-6 lg:-mt-10">
              <Image
                src="/Connextion3.png"
                alt="ConNextioN"
                width={560}
                height={560}
                priority
                className="w-full max-w-md lg:max-w-lg h-auto"
              />
            </div>
          </div>          
        </div>
        
      </section>

      {/* Features */}
      <section className="mt-24">
        <div className="max-w-7xl mx-auto pl-4 md:pl-10 pr-6 pt-24 pb-20 md:pt-28 md:pb-24">
          <div className="grid md:grid-cols-3 gap-x-12 gap-y-16 pt-12">
            <div>
              <div className="text-md font-lg tracking-wider text-neutral-500 mb-4">
                Direct Connection
              </div>
              <h3 className="text-xl font-semibold mb-3">
                Connect with anyone, anywhere
              </h3>
              <p className="text-neutral-600 leading-relaxed">
                Not only private chats but also group chat live at real time.
              </p>
            </div>

            <div>
              <div className="text-xs font-medium tracking-wider text-neutral-500 mb-4">
                Privacy matters
              </div>
              <h3 className="text-xl font-semibold mb-3">Your life, your way</h3>
              <p className="text-neutral-600 leading-relaxed">
                Control the level of privacy for your profile, posts, and chats.
              </p>
            </div>

            <div>
              <div className="text-xs font-medium tracking-wider text-neutral-500 mb-4">
                Events and Groups
              </div>
              <h3 className="text-xl font-semibold mb-3">Choose your circle</h3>
              <p className="text-neutral-600 leading-relaxed">
                Share updates, photos, create groups and events 
                <br />
                and many more with
                <br />
                your community or just with your friends.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="bg-neutral-50 mt-24">
        <div className="max-w-7xl mx-auto pl-4 md:pl-10 pr-6 py-20 md:py-28">
          <div className="max-w-4xl">
            <h2 className="text-4xl md:text-5xl font-semibold leading-tight tracking-tight mb-6">
              Ready to connect
              <br />
              and explore?
            </h2>
            <p className="text-lg md:text-xl text-neutral-600 mb-10 max-w-2xl">
              Create your account, create your new ConNextion.
            </p>

            <Link
              href="/register"
              className="inline-block text-base font-medium text-neutral-900 underline underline-offset-8 decoration-2 decoration-neutral-900/30 hover:decoration-neutral-900 transition"
            >
              Create a new account
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer>
        <div className="max-w-7xl mx-auto pl-4 md:pl-10 pr-6 py-10">
          <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-8">
            <div className="text-sm text-neutral-500">
              (c) {new Date().getFullYear()} ConNextioN
            </div>
            <div className="flex gap-8 text-sm font-medium">
              <Link href="#" className="text-neutral-600 hover:text-neutral-900 transition">
                Privacy
              </Link>
              <Link href="#" className="text-neutral-600 hover:text-neutral-900 transition">
                Terms
              </Link>
              <Link href="#" className="text-neutral-600 hover:text-neutral-900 transition">
                Contact
              </Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
