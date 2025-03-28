import { initializeApp } from "firebase/app";
import { getAuth, GoogleAuthProvider, signInWithPopup, getAdditionalUserInfo } from "firebase/auth";
import { collection, doc, getDoc, getFirestore, onSnapshot } from "firebase/firestore";
import { getAnalytics } from "firebase/analytics";

const env = import.meta.env;
const config = {
    apiKey: env.VITE_FIREBASE_API_KEY,
    authDomain: env.VITE_FIREBASE_AUTH_DOMAIN,
    databaseURL: env.VITE_FIREBASE_DATABASE_URL,
    projectId: env.VITE_FIREBASE_PROJECT_ID,
    storageBucket: env.VITE_FIREBASE_STORAGE_BUCKET,
    messagingSenderId: env.VITE_FIREBASE_MESSAGING_SENDER_ID,
    appId: env.VITE_FIREBASE_APP_ID,
    measurementId: env.VITE_FIREBASE_MEASUREMENT_ID
};

class Firebase {
    constructor() {
        const app = initializeApp(config);

        /* Firebase APIs */
        getAnalytics(app);
        this.auth = getAuth(app);
        this.firestore = getFirestore(app);

        /* Social Sign In Method Provider */
        this.googleProvider = new GoogleAuthProvider();
    }

    // *** Auth API ***

    doSignInWithGoogle() {
        return signInWithPopup(this.auth, this.googleProvider);
    }

    getAdditionalUserInfo = user => getAdditionalUserInfo(user);

    doSignOut = () => this.auth.signOut();

    // *** Merge Auth and DB User API *** //
    onAuthUserListener = (cb, fallback) => {
        this.auth.onAuthStateChanged(async authUser => {
            if (authUser) {
                const userRef = doc(this.firestore, "users", authUser.uid);
                const userSnap = await getDoc(userRef);

                let dbUser = undefined;
                if (userSnap.exists()) {
                    dbUser = userSnap.data();

                    // Default empty roles
                    if (!dbUser.roles) {
                        dbUser.roles = [];
                    }
                } else {
                    console.log(`No such user[${authUser.uid}] found!`);
                }

                // Merge auth and db user
                const mergedUser = {
                    uid: authUser.uid,
                    email: authUser.email,
                    emailVerified: authUser.emailVerified,
                    providerData: authUser.providerData,
                    ...dbUser
                };

                cb(mergedUser);
            } else {
                fallback();
            }
        });
    };

    onUserListener = (uid, cb) => {
        const userRef = doc(this.firestore, "users", uid);
        return onSnapshot(userRef, docSnap => {
            cb(docSnap.exists() ? docSnap.data() : null);
        });
    };

    // *** User API ***

    user = uid => doc(this.firestore, "users", uid);
    users = () => collection(this.firestore, "users");

    // *** Message API ***

    // message = uid => this.firestore.ref(`messages/${uid}`);

    // messages = () => this.firestore.ref('messages');
}

export default Firebase;
