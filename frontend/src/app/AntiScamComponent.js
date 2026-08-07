import { DCLogic } from '../runtime/DCLogic.js'
import { React } from '../runtime/vdom.js'

const _LS = 'antiscam.v1'

export class AntiScamComponent extends DCLogic {
  CAT = {
    pay: 'Payment Issues',
    ship: 'Shipping Scams',
    fake: 'Fake Items',
    acc: 'Account & Profile',
    off: 'Off-Platform',
    ret: 'Returns',
  }

  MISSIONS = [
    {
      id: 'too-good',
      t: 'Too Good to Be True',
      d: 'A brand new console for half the price?',
      dif: 'EASY',
      role: 'BUYER',
      xp: 10,
      cat: 'fake',
      ic: 'console',
      feat: 1,
      play: 1,
      mins: 5,
      new: 0,
    },
    {
      id: 'fake-shipping',
      t: 'Fake Shipping Info',
      d: 'Suspicious tracking or wrong address?',
      dif: 'MEDIUM',
      role: 'SELLER',
      xp: 20,
      cat: 'ship',
      ic: 'truck',
      feat: 1,
      play: 1,
      mins: 6,
      new: 0,
    },
    {
      id: 'payment-outside',
      t: 'Payment Outside App',
      d: 'They want to take the chat somewhere else.',
      dif: 'HARD',
      role: 'BOTH',
      xp: 15,
      cat: 'pay',
      ic: 'card',
      feat: 1,
      mins: 7,
      new: 0,
    },
    {
      id: 'damaged-claim',
      t: 'Damaged Item Claim',
      d: 'Buyer claims the item arrived damaged.',
      dif: 'MEDIUM',
      role: 'SELLER',
      xp: 15,
      cat: 'ret',
      ic: 'box',
      new: 1,
      mins: 6,
    },
    {
      id: 'account-takeover',
      t: 'Account Takeover',
      d: 'Strange login activity detected.',
      dif: 'HARD',
      role: 'BOTH',
      xp: 20,
      cat: 'acc',
      ic: 'lock',
      new: 1,
      mins: 8,
    },
    {
      id: 'chat-red-flags',
      t: 'Chat Red Flags',
      d: 'Spot the suspicious messages.',
      dif: 'EASY',
      role: 'BOTH',
      xp: 10,
      cat: 'off',
      ic: 'chat',
      new: 1,
      mins: 4,
    },
    {
      id: 'fake-prize',
      t: 'Fake Prize Giveaway',
      d: 'Looks real, but it is a scam.',
      dif: 'MEDIUM',
      role: 'BUYER',
      xp: 15,
      cat: 'fake',
      ic: 'gift',
      new: 1,
      mins: 5,
    },
    {
      id: 'keen-eye',
      t: 'Keen Eye',
      d: 'Spot the red flags like a pro.',
      dif: 'EASY',
      role: 'BUYER',
      xp: 15,
      cat: 'fake',
      ic: 'glass',
      mins: 5,
      arch: 1,
    },
    {
      id: 'urgent-payment',
      t: 'Urgent Payment',
      d: 'Pressure tactics to rush you.',
      dif: 'HARD',
      role: 'BOTH',
      xp: 20,
      cat: 'pay',
      ic: 'clock',
      lock: 'Reach Level 8 to unlock this mission.',
      mins: 7,
      arch: 1,
    },
    {
      id: 'qr-trap',
      t: 'QR Code Trap',
      d: 'Do not get scanned the wrong way.',
      dif: 'MEDIUM',
      role: 'BUYER',
      xp: 15,
      cat: 'pay',
      ic: 'qr',
      lock: 'Complete “Payment Outside App” to unlock.',
      mins: 6,
      arch: 1,
    },
    {
      id: 'phishing-links',
      t: 'Phishing Links',
      d: 'Not every link is what it seems.',
      dif: 'EASY',
      role: 'BOTH',
      xp: 10,
      cat: 'off',
      ic: 'link',
      lock: 'Complete 20 missions to unlock.',
      mins: 4,
      arch: 1,
    },
    {
      id: 'fake-support',
      t: 'Fake Support Agent',
      d: '“Support” asks for your login code.',
      dif: 'MEDIUM',
      role: 'SELLER',
      xp: 15,
      cat: 'acc',
      ic: 'shield',
      mins: 6,
    },
  ]

  SCRIPTS = {
    'too-good': {
      title: 'Too Good to Be True',
      role: 'BUYER',
      with: 'SELLER',
      conv: '#7F3B2A',
      obj: [
        'Keep the conversation going and identify red flags.',
        'Hold the line — do not move money off-platform.',
        'Decide what to do with the link you were sent.',
      ],
      steps: [
        {
          who: 'seller',
          t: "Hey! I saw you're interested in the PS5. It's still available!",
          time: '10:32 AM',
        },
        { who: 'you', t: "Hi! Yes, I'm interested. Is it still brand new?", time: '10:33 AM' },
        {
          who: 'seller',
          t: "Yep, brand new, sealed. I can ship it to you today. I just need payment through PayFriends (friends & family) so I don't get hit with fees.",
          time: '10:34 AM',
        },
        {
          d: 1,
          risk: 'medium',
          skill: 'Platform Safety',
          sigs: [
            { tone: 'red', ic: 'flagr', t: 'Asking for payment outside platform' },
            {
              tone: 'gold',
              ic: 'warn',
              t: 'Too good to be true',
              s: '(price is $250 below market)',
            },
          ],
          opts: [
            {
              k: 'safe',
              t: "Why not use the marketplace checkout? It's safer for both of us.",
              log: 'Kept payment on the platform',
              note: 'You refused an off-platform transfer.',
              fb: {
                what: 'You pushed the deal back to marketplace checkout. The seller lost the quiet exit they were aiming for.',
                why: 'Checkout holds the money until the item is delivered. A “friends & family” transfer has no buyer protection and is almost impossible to reverse.',
                sign: 'A seller who blames “fees” to move you onto a payment method with no protection.',
                real: 'Pay inside the app, every time. If the seller refuses, end the deal and report the listing.',
              },
            },
            {
              k: 'warn',
              t: "Can you prove it's new? Any photos?",
              log: 'Asked for more photos',
              note: 'Reasonable, but it does not address the payment risk.',
              fb: {
                what: 'You asked for proof of the item. The seller will happily send photos — they cost nothing.',
                why: 'Photos can be copied from anywhere. Checking the item does not make an unprotected payment safe, so the real risk is still on the table.',
                sign: 'Attention drifting to the product while the payment method stays unsafe.',
                real: 'Verify the item, but never let it distract you from how the money moves.',
              },
            },
            {
              k: 'risk',
              t: 'Okay, send me the PayFriends details.',
              log: 'Agreed to pay off-platform',
              note: 'The money would leave the protected checkout.',
              fb: {
                what: 'You agreed to send money outside the marketplace. At this point the platform can no longer see, hold, or refund the payment.',
                why: '“Friends & family” transfers are designed for people you trust. There is no dispute process and no seller verification behind them.',
                sign: 'Any request to pay in a way that removes buyer protection.',
                real: 'If someone insists on an unprotected transfer, stop. That single request is enough to walk away.',
              },
            },
          ],
        },
        {
          who: 'seller',
          t: "The fee's way too high and slows things down. I'm a legit seller, check my reviews!",
          time: '10:36 AM',
        },
        {
          who: 'seller',
          t: "Also I have 3 other buyers waiting. If you send it in the next 20 minutes it's yours, otherwise I move on.",
          time: '10:37 AM',
        },
        {
          d: 2,
          risk: 'high',
          skill: 'Pressure Resistance',
          sigs: [
            { tone: 'red', ic: 'skull', t: 'Creating urgency', s: '(“other buyers waiting”)' },
          ],
          opts: [
            {
              k: 'warn',
              t: 'Can you hold it for an hour? I need to think.',
              log: 'Asked for more time',
              note: 'Buying time helps, but the risk is unchanged.',
              fb: {
                what: 'You asked for time. The seller may agree — then keep pushing with a new deadline.',
                why: 'Delaying a bad payment is not the same as refusing it. The unprotected transfer is still the plan.',
                sign: 'A countdown that keeps resetting whenever you hesitate.',
                real: 'Use the pause to check the profile and the price — then decide on the method, not the clock.',
              },
            },
            {
              k: 'safe',
              t: "I'm not comfortable paying outside the platform.",
              log: 'Held the line under pressure',
              note: 'You did not let the deadline decide for you.',
              fb: {
                what: 'You named the problem and stayed with it. Deadlines only work on people who feel rushed.',
                why: 'Artificial urgency exists to stop you from checking things. A genuine seller loses nothing by using checkout.',
                sign: '“Other buyers”, countdowns, and “decide now” — pressure aimed at your judgement, not at the item.',
                real: 'Slow the conversation down on purpose. If that ends the deal, the deal was never real.',
              },
            },
            {
              k: 'risk',
              t: "Fine — sending now so I don't lose it.",
              log: 'Rushed the payment',
              note: 'Urgency drove the decision.',
              fb: {
                what: 'You let the deadline decide. Money sent this way usually cannot be recovered.',
                why: 'Urgency plus an unprotected payment method is the classic combination in marketplace fraud — it removes your time to verify.',
                sign: 'A deadline appearing right after you raised a concern.',
                real: 'When you feel rushed, that is the signal to stop, not to hurry.',
              },
            },
          ],
        },
        {
          who: 'sys',
          t: 'This conversation contains a link that leaves the marketplace.',
          time: '10:38 AM',
        },
        {
          who: 'seller',
          t: "No problem, use this secure page instead: pay-secure-market.net/ps5 — it's the official one.",
          time: '10:39 AM',
        },
        {
          d: 3,
          risk: 'critical',
          skill: 'Link Safety',
          sigs: [
            {
              tone: 'red',
              ic: 'link',
              t: 'Look-alike payment domain',
              s: '(pay-secure-market.net)',
            },
          ],
          opts: [
            {
              k: 'risk',
              t: 'Opening it now.',
              log: 'Opened an unverified link',
              note: 'The page was a look-alike payment form.',
              fb: {
                what: 'You opened a page built to look like marketplace checkout. Anything typed there — card number, code, password — goes straight to the scammer.',
                why: 'The domain is not the marketplace. Look-alike names (extra words, different endings) are the cheapest part of a scam to fake.',
                sign: 'A payment page that arrives as a chat link instead of from the app itself.',
                real: 'Never pay through a link someone sent you. Open the app yourself and pay from the order.',
              },
            },
            {
              k: 'warn',
              t: "Is that link official? I'll check with support.",
              log: 'Questioned the link',
              note: 'Right instinct — but do not ask the person sending it.',
              fb: {
                what: 'You paused before clicking. Good — but the seller will confirm their own link is “official”.',
                why: 'Verification only counts when it comes from a source the scammer does not control. Ask the platform, not the counterpart.',
                sign: 'Being invited to verify with the same person who sent the link.',
                real: 'Open support from inside the app. Never through a link, number, or contact given to you in chat.',
              },
            },
            {
              k: 'safe',
              t: "That domain isn't the marketplace. I'll report this and pay through checkout.",
              log: 'Refused the phishing link',
              note: 'You checked the domain before clicking.',
              fb: {
                what: 'You read the domain, recognised it as fake, and reported it instead of clicking.',
                why: 'The real checkout never lives on a separate domain sent in chat. Reading the address bar takes two seconds and stops most payment phishing.',
                sign: 'Payment domains with extra words like “secure”, “pay”, or hyphenated brand names.',
                real: 'Check the domain, then pay from inside the app. Report the message so the listing gets reviewed.',
              },
            },
          ],
        },
      ],
    },
    'fake-shipping': {
      title: 'Fake Shipping Info',
      role: 'SELLER',
      with: 'BUYER',
      conv: '#2A4B7F',
      obj: [
        'Agree delivery terms without exposing your details.',
        'Verify who is really messaging you.',
        'Confirm the payment before anything ships.',
      ],
      steps: [
        { who: 'seller', t: "Hi! I'll take the bike. Can you ship it today?", time: '09:12 AM' },
        {
          who: 'you',
          t: "Sure — I'll book the pickup through marketplace delivery.",
          time: '09:14 AM',
        },
        {
          who: 'seller',
          t: "My courier is way cheaper. Send me your card number and I'll transfer the shipping cost right now.",
          time: '09:15 AM',
        },
        {
          d: 1,
          risk: 'medium',
          skill: 'Payment Safety',
          sigs: [
            { tone: 'red', ic: 'flagr', t: 'Asking for card details in chat' },
            { tone: 'gold', ic: 'warn', t: 'Pushing an unknown courier' },
          ],
          opts: [
            {
              k: 'safe',
              t: 'Delivery is arranged through the platform only — no card details in chat.',
              log: 'Protected payment details',
              note: 'You kept your card out of the conversation.',
              fb: {
                what: 'You kept both the delivery and your card inside the platform, where the order is tracked.',
                why: 'A buyer never needs your card number to send you money. Card details in chat are the start of most seller-side fraud.',
                sign: 'Any request for card numbers, CVV, or bank credentials “to send you money”.',
                real: 'Share nothing but the order number. Payouts arrive through the platform balance.',
              },
            },
            {
              k: 'warn',
              t: 'Which courier company do you use?',
              log: 'Asked about the courier',
              note: 'Polite, but it keeps the unsafe offer alive.',
              fb: {
                what: 'You engaged with the courier idea. The buyer will invent a plausible name and keep going.',
                why: 'The problem is not which courier — it is that the delivery and the payment would leave the platform’s protection at once.',
                sign: 'A negotiation that quietly moves both money and logistics off-platform.',
                real: 'Keep delivery inside the marketplace so the tracking and payout stay linked to the order.',
              },
            },
            {
              k: 'risk',
              t: "Sure, here's my card number.",
              log: 'Shared card details',
              note: 'Card data was exposed in chat.',
              fb: {
                what: 'You handed over card details. That data can be reused, sold, or used for verification attempts against your account.',
                why: 'Receiving money never requires your full card number. Anyone asking for it wants more than shipping costs.',
                sign: '“I need your card to pay you” — the request itself is the red flag.',
                real: 'Never send card, CVV, or bank credentials in a chat. Payouts always go through the platform.',
              },
            },
          ],
        },
        { who: 'sys', t: 'New message from “Marketplace Delivery Support”.', time: '09:21 AM' },
        {
          who: 'support',
          t: 'Hello! To release the payment for order #4471 we need the confirmation code we just sent to your phone.',
          time: '09:22 AM',
        },
        {
          d: 2,
          risk: 'high',
          skill: 'Account Safety',
          sigs: [
            { tone: 'red', ic: 'skull', t: 'Support impersonation' },
            { tone: 'red', ic: 'flagr', t: 'Asking for an SMS code' },
          ],
          opts: [
            {
              k: 'warn',
              t: 'Which order? I have several going right now.',
              log: 'Asked for order details',
              note: 'You stayed in the conversation instead of leaving it.',
              fb: {
                what: 'You asked a clarifying question. The impersonator now has a thread to keep pulling.',
                why: 'Every reply gives them another chance to sound official. Real support does not need you to prove anything by SMS code.',
                sign: '“Support” that appears in a chat you did not open from the app.',
                real: 'Close the conversation and open the order in the app. Contact support only from inside it.',
              },
            },
            {
              k: 'risk',
              t: 'The code is 883-104.',
              log: 'Shared an SMS code',
              note: 'That code could authorise a login or transfer.',
              fb: {
                what: 'You gave away a one-time code. Those codes confirm logins, password resets, and payouts.',
                why: 'Anyone holding the code can act as you. That is exactly why the message asking for it always sounds urgent and official.',
                sign: 'A code request arriving right when you were expecting money.',
                real: 'Never share SMS codes — not with support, couriers, or buyers. No legitimate service asks for them.',
              },
            },
            {
              k: 'safe',
              t: "Support never asks for SMS codes. I'll check the order in the app.",
              log: 'Refused to share the SMS code',
              note: 'You verified through the app instead.',
              fb: {
                what: 'You refused the code and went to the app to check the order yourself.',
                why: 'One-time codes are the last line of defence on your account. Real support has your order status already — they never need your code.',
                sign: 'Anyone at all asking you to read out a code from an SMS.',
                real: 'Verify from inside the app. If a message claims to be support, treat the message as fake until the app agrees.',
              },
            },
          ],
        },
        {
          who: 'seller',
          t: "I already paid! Here's the receipt screenshot. Ship it now or I'll report you.",
          time: '09:31 AM',
        },
        {
          d: 3,
          risk: 'high',
          skill: 'Fraud Instinct',
          sigs: [{ tone: 'gold', ic: 'warn', t: 'Payment “proof” as a screenshot' }],
          opts: [
            {
              k: 'safe',
              t: "I'll ship as soon as the payment shows in my marketplace balance.",
              log: 'Waited for real confirmation',
              note: 'You trusted the platform, not a screenshot.',
              fb: {
                what: 'You made shipping depend on the balance in your account, not on an image in a chat.',
                why: 'Screenshots are trivial to fake or edit. The only proof that matters is the status the platform itself shows.',
                sign: 'Payment “confirmed” by an image, plus a threat to make you act fast.',
                real: 'Ship only after the order shows as paid in the app. A real buyer can wait a few minutes.',
              },
            },
            {
              k: 'warn',
              t: 'Can you send the transaction ID?',
              log: 'Asked for a transaction ID',
              note: 'Better than a screenshot, but still their evidence.',
              fb: {
                what: 'You asked for a stronger-looking proof. An ID can be invented just as easily as a screenshot.',
                why: 'Any evidence supplied by the other side can be fabricated. Only your own account view is reliable.',
                sign: 'Proof of payment that always comes from the buyer instead of the platform.',
                real: 'Check your marketplace balance or order status directly — that is the only receipt that counts.',
              },
            },
            {
              k: 'risk',
              t: 'Okay, shipping today.',
              log: 'Shipped before payment cleared',
              note: 'The item left before the money existed.',
              fb: {
                what: 'You shipped on the strength of a screenshot and a threat. The payment never arrives.',
                why: 'Combining fake proof with a report threat is designed to make you act before you check. It targets sellers who fear bad ratings.',
                sign: 'Threats and deadlines attached to unverified payment proof.',
                real: 'Never dispatch before the platform shows the order as paid. Ratings can be disputed; a shipped item cannot be returned.',
              },
            },
          ],
        },
      ],
    },
  }

  IC = {
    shield: [
      '..aaaaaaaa../.abbbbbbbba./abbbbbbbbbba/abbbbbbbccba/abbbbbbcccba/abbcbbcccbba/abbccccccbba/.abbccccbba./.abbbccbbba./..abbbbbba../...abbbba.../....aaaa....',
      { a: '#0B6B4F', b: '#2FE0A0', c: '#06231A' },
    ],
    shieldw: [
      '..aaaaaaaa../.abbbbbbbba./abbbbbbbbbba/abbbbbbbccba/abbbbbbcccba/abbcbbcccbba/abbccccccbba/.abbccccbba./.abbbccbbba./..abbbbbba../...abbbba.../....aaaa....',
      { a: '#1E5FB8', b: '#8FC9FF', c: '#0B2444' },
    ],
    flag: [
      '.aa........./.aabbbbbb.../.aabbbbbbbb./.aabbbbbbbb./.aabbbbbb.../.aa........./.aa........./.aa........./.aa........./.aaaa.......',
      { a: '#1E6B58', b: '#2FE0A0' },
    ],
    flagr: [
      '.aa........./.aabbbbbb.../.aabbbbbbbb./.aabbbbbb.../.aa........./.aa........./.aa........./.aaaa.......',
      { a: '#8C1A18', b: '#F0524B' },
    ],
    bars: [
      '............/........aa../....aa..aa../....aa..aa../.aa.aa..aa../.aa.aa..aa../.aa.aa..aa../.aa.aa..aa../.aa.aa..aa../............',
      { a: '#4FD8E8' },
    ],
    book: [
      '............/.aaaaaaaaaa./.abbbbabbba./.abbbbabbba./.abbbbabbba./.abbbbabbba./.abbbbabbba./.aaaaaaaaaa./............',
      { a: '#1E5FB8', b: '#9EC5FF' },
    ],
    gear: [
      '...aa..aa.../...aa..aa.../.aaaaaaaaaa./.aaa....aaa./aaa......aaa/aaa......aaa/aaa......aaa/aaa......aaa/.aaa....aaa./.aaaaaaaaaa./...aa..aa.../...aa..aa...',
      { a: '#8FA3BD' },
    ],
    target: [
      '...aaaaaa.../..aa....aa../.aa..bb..aa./aa..bbbb..aa/aa.bb..bb.aa/aa.bb..bb.aa/aa..bbbb..aa/aa...bb...aa/.aa......aa./..aa....aa../...aaaaaa...',
      { a: '#F0524B', b: '#2FE0A0' },
    ],
    console: [
      '..aaaaaaaa../.abbbbbbbba./.abccccccba./.abccccccba./.abbbbbbbba./.abbbbbbbba./.abdbbbbdba./.abbbbbbbba./.abdbdbdbba./.abbbbbbbba./..aaaaaaaa..',
      { a: '#4A2E7A', b: '#A97BFF', c: '#2A1750', d: '#E4D6FF' },
    ],
    truck: [
      '............/.bbbbbb...../.bbbbbbaaaa./.bbbbbbaaaa./.bbbbbbbbbb./.bbbbbbbbbb./.bbbbbbbbbb./..cc....cc../.cccc..cccc./..cc....cc..',
      { b: '#F5A03A', a: '#FFD08A', c: '#2B3A4E' },
    ],
    card: [
      '............/.aaaaaaaaaa./.abbbbbbbba./.acccccccca./.abbbbbbbba./.abbbbbbbba./.abddbbbbba./.abbbbbbbba./.aaaaaaaaaa./............',
      { a: '#1E5FB8', b: '#3D8BFD', c: '#0E2A55', d: '#CFE4FF' },
    ],
    box: [
      '............/.aaaaaaaaaa./.abbbcbbbba./.abbbcbbbba./.acccccccca./.abbbcbbbba./.abbbcbbbba./.abbbcbbbba./.aaaaaaaaaa./............',
      { a: '#7A4F22', b: '#C88A45', c: '#E8C089' },
    ],
    lock: [
      '...aaaaaa.../..aa....aa../..aa....aa../.bbbbbbbbbb./.bbbbbbbbbb./.bbbbccbbbb./.bbbbccbbbb./.bbbbbbbbbb./.bbbbbbbbbb./............',
      { a: '#8FA3BD', b: '#B8C8DC', c: '#3A4E66' },
    ],
    chat: [
      '.aaaaaaaaaa./.abbbbbbbba./.abccccccba./.abbbbbbbba./.abccccbbba./.abbbbbbbba./.aaaaaaaaaa./...aa......./..aa......../............',
      { a: '#1E5FB8', b: '#3D8BFD', c: '#CFE4FF' },
    ],
    gift: [
      '..b......b../...bb..bb.../.aaaaccaaaa./.aaaaccaaaa./.dddddddddd./.aaaaccaaaa./.aaaaccaaaa./.aaaaccaaaa./.aaaaccaaaa./............',
      { a: '#E05A52', b: '#F5C84A', c: '#F5C84A', d: '#F5C84A' },
    ],
    glass: [
      '..aaaa....../.abbbba...../abbbbbba..../abbbbbba..../abbbbbba..../.abbbba...../..aaaa.a..../.......aa.../........aa../.........aa.',
      { a: '#4FD8E8', b: '#0E2A38' },
    ],
    clock: [
      '...aaaa...../..abbbba..../.abbcbbba.../.abbcbbba.../.abbccbba.../.abbbbbba.../..abbbba..../...aaaa.....',
      { a: '#F5A03A', b: '#FFD9A0', c: '#7A4F22' },
    ],
    qr: [
      'aaaa..aaaa../a..a..a..a../a..a..a..a../aaaa..aaaa../............/aaaa..a.a.a./a..a..aa.aa./a..a..a.a.a./aaaa..aa.aa.',
      { a: '#D6E5F5' },
    ],
    link: [
      '....aaaa..../...a....a.../..a......a../..a..bb..a../...a.bb.a.../.....bb...../...a.bb.a.../..a......a../..a......a../...aaaaaa...',
      { a: '#A97BFF', b: '#C7A8FF' },
    ],
    trophy: [
      'a..aaaaaa..a/a.abbbbbba.a/aaabbbbbbaaa/.aabbbbbbaa./..abbbbbba../...abbbba.../....abba..../....abba..../..aaaaaaaa../..aaaaaaaa..',
      { a: '#C9902A', b: '#F5C84A' },
    ],
    fire: [
      '.....aa...../....abba..../...abbbba.../..abbccbba../.abbccccbba./.abccccccba./abccccccccba/abcccddcccba/.abccddccba./..abbbbbba..',
      { a: '#C93A18', b: '#F0672A', c: '#F5A03A', d: '#FFE08A' },
    ],
    skull: [
      '..aaaaaaaa../.aaaaaaaaaa./aabbaaaabbaa/aabbaaaabbaa/aaaaaaaaaaaa/aaaabbaaaaaa/.aaaaaaaaaa./..aaaaaaaa../..a.a.a.a.a.',
      { a: '#F0524B', b: '#2A0C0C' },
    ],
    star: [
      '.....aa...../.....aa...../....abba..../aaaabbbbaaaa/.aabbbbbbaa./..abbbbbba../.abbbaabbba./abba....abba',
      { a: '#C9902A', b: '#F5C84A' },
    ],
    heart: [
      '.aaa.aaa../abbbabbba/abbbbbbbb/abbbbbbbb/.abbbbbbb/..abbbbb./...abbb../....ab.../.........',
      { a: '#8C1A18', b: '#F0524B' },
    ],
    bulb: [
      '...aaaa.../..abbbba../.abbbbbba./.abbbbbba./.abbbbbba./..abbbba../...acca.../...acca.../....cc....',
      { a: '#C9902A', b: '#FFE08A', c: '#8FA3BD' },
    ],
    warn: [
      '....aa..../....aa..../...abba.../...abba.../..abccba../..abccba../.abbccbba./.abbbbbba./aabbccbbaa/aaaaaaaaaa',
      { a: '#C98A18', b: '#F5C84A', c: '#3A2A08' },
    ],
    sound: [
      '....aa..../...aaa..b./.aaaaa.b.b/aaaaaa.b.b/aaaaaa.b.b/.aaaaa.b.b/...aaa..b./....aa....',
      { a: '#2FE0A0', b: '#2FE0A0' },
    ],
    mute: [
      '....aa..../...aaa..../.aaaaa.b.b/aaaaaa..b./aaaaaa.b.b/.aaaaa..../...aaa..../....aa....',
      { a: '#5E7794', b: '#F0524B' },
    ],
    medal: [
      '...aaaa.../..abbbba../.abbccbba./.abcccbba./.abbccbba./..abbbba../...aaaa.../..a.aa.a../.a..aa..a.',
      { a: '#C9902A', b: '#F5C84A', c: '#8C6510' },
    ],
    chest: [
      '.aaaaaaaa./abbbbbbbba/abcccccccba'.replace('ba/abcccccccba', 'ba/abccccccba'),
      { a: '#7A4F22', b: '#C88A45', c: '#F5C84A' },
    ],
    cart: [
      '..a......./..aaaaaaa./..abbbbba./..abbbbba./..aaaaaaa./...a...a../..aaa.aaa./...a...a..',
      { a: '#2FE0A0', b: '#0E3A2C' },
    ],
    store: [
      'aaaaaaaaaa/abbbbbbbba/aaaaaaaaaa/abbbbbbbba/abbaaaabba/abbaccabba/abbaccabba/aaaaaaaaaa',
      { a: '#E05A52', b: '#FFB4AE', c: '#7A1A16' },
    ],
    check: [
      '........./.......a./......aa./.a...aa../.aa.aa.../..aaa..../...a...../.........',
      { a: '#2FE0A0' },
    ],
    wizard: [
      '.....vvvvvv...../....vvvvvvvv..../...vvvvvvvvvv.../...vvssssssvv.../...vseesseesv.../...vssssssssv.../....ssmmmmss..../...vvssssssvv.../..vvvvvvvvvvvv../..vvvvvvvvvvpp../..vvvvvvvvvvpP../...VVVVVVVVVV...',
      {
        v: '#7B4FD0',
        V: '#55379E',
        s: '#F2C9A0',
        e: '#22150E',
        m: '#C97B54',
        p: '#29384C',
        P: '#9FD8FF',
      },
    ],
    buyer: [
      '.....cccccc...../....cccccccc..../...cccccccccc.../...CCCCCCCCCC.../...hhsssssshh.pp/...hseesseesh.pP/...hssssssssh.pP/....ssmmmmss..pp/.....ssssss..ss./...jjjjjjjjjjjs./..jJlljjjjlljJ../..jjlljjjjlljj../..jjjjjjjjjjjj../..jjjjjjjjjjjj../..JJJJJJJJJJJJ../...dd......dd...',
      {
        c: '#23B884',
        C: '#17805E',
        h: '#2A1B12',
        s: '#F2C9A0',
        e: '#22150E',
        m: '#C97B54',
        j: '#1E9E74',
        J: '#12684C',
        l: '#7CFFD0',
        p: '#29384C',
        P: '#9FD8FF',
        d: '#24334A',
      },
    ],
    seller: [
      '.....kkkkkk...../....kkkkkkkk..../...kkkkkkkkkk.../...KKKKKKKKKK.../...hhsssssshh.../...hseesseesh.../...hssssssssh.../....ssmmmmss..../.....ssssss...../...jjjjjjjjjj.../..jjjjjjjjjjjj../..jsbbbbbbbbsj../..jsbbbttbbbsj../..jsbbbttbbbsj../..jjBBBBBBBBjj../...dd......dd...',
      {
        k: '#E04840',
        K: '#A8302A',
        h: '#2A1B12',
        s: '#F2C9A0',
        e: '#22150E',
        m: '#C97B54',
        j: '#C23A34',
        b: '#C88A45',
        B: '#8A5526',
        t: '#E8C089',
        d: '#24334A',
      },
    ],
  }

  state = (() => {
    let s = {}
    try {
      s = JSON.parse(localStorage.getItem(_LS)) || {}
    } catch (e) {
      s = {}
    }
    return Object.assign(
      {
        route: 'home',
        q: '',
        fRole: 'ALL',
        fDif: [],
        fCat: null,
        fProg: 'all',
        sort: 'dif',
        sound: true,
        vol: 0.35,
        motion: 'standard',
        profile: false,
        modal: null,
        toast: null,
        busy: null,
        xp: 540,
        streak: 5,
        completed: ['keen-eye'],
        badgeCount: 18,
        mid: null,
        ci: 0,
        msgs: [],
        typing: false,
        dec: null,
        sel: null,
        fb: null,
        sigs: [],
        risk: 'low',
        log: [],
        gained: 0,
        elapsed: 0,
        result: null,
        scores: {},
        tipsOpen: false,
        ruleCat: 'all',
        ruleQ: '',
      },
      s.saved || {}
    )
  })()

  // ---------- lifecycle ----------
  componentDidMount() {
    this.chatRef = React.createRef()
    this.onHash = () => {
      const h = (location.hash || '').replace(/^#\/?/, '')
      const route = h.split('?')[0] || 'home'
      const known = [
        'home',
        'missions',
        'scenarios',
        'play',
        'result',
        'progress',
        'rules',
        'settings',
      ]
      if (known.indexOf(route) >= 0) {
        if (route !== this.state.route) this.setState({ route, profile: false })
      } else if (route) this.setState({ route: 'nf', profile: false })
    }
    window.addEventListener('hashchange', this.onHash)
    this.onKey = (e) => {
      if (e.key === 'Escape' && (this.state.modal || this.state.profile))
        this.setState({ modal: null, profile: false })
    }
    document.addEventListener('keydown', this.onKey)
    this.onDocClick = (e) => {
      if (
        this.state.profile &&
        this.profRef &&
        this.profRef.current &&
        !this.profRef.current.contains(e.target)
      )
        this.setState({ profile: false })
    }
    document.addEventListener('mousedown', this.onDocClick)
    this.profRef = React.createRef()
    this.onResize = () => this.setState({ vw: window.innerWidth })
    window.addEventListener('resize', this.onResize)
    const p = this.props || {},
      patch = { vw: window.innerWidth }
    if (p.soundEnabled != null) patch.sound = !!p.soundEnabled
    if (p.motionMode === 'standard' || p.motionMode === 'reduced') patch.motion = p.motionMode
    const h = (location.hash || '').replace(/^#\/?/, '').split('?')[0]
    if (h)
      patch.route =
        [
          'home',
          'missions',
          'scenarios',
          'play',
          'result',
          'progress',
          'rules',
          'settings',
        ].indexOf(h) >= 0
          ? h
          : 'nf'
    else if (p.startScreen) patch.route = p.startScreen
    if (patch.route === 'play' && (!this.state.mid || this.state.result)) patch.route = 'missions'
    if (patch.route === 'result' && !this.state.result) patch.route = 'missions'
    this.setState(patch, () => {
      if (
        this.state.route === 'play' &&
        this.state.mid &&
        !this.state.result &&
        !this.state.fb &&
        !this.state.dec
      )
        this.advance()
    })
    this.timer = setInterval(() => {
      if (this.state.route === 'play' && !this.state.result)
        this.setState((s) => ({ elapsed: s.elapsed + 1 }))
    }, 1000)
  }
  componentWillUnmount() {
    window.removeEventListener('hashchange', this.onHash)
    window.removeEventListener('resize', this.onResize)
    document.removeEventListener('keydown', this.onKey)
    document.removeEventListener('mousedown', this.onDocClick)
    clearInterval(this.cnt)
    clearInterval(this.timer)
    if (this.toastT) clearTimeout(this.toastT)
  }
  componentDidUpdate() {
    const el = this.chatRef && this.chatRef.current
    if (el) el.scrollTop = el.scrollHeight
  }

  persist(extra) {
    const s = Object.assign({}, this.state, extra || {})
    const saved = {
      xp: s.xp,
      streak: s.streak,
      completed: s.completed,
      badgeCount: s.badgeCount,
      sound: s.sound,
      vol: s.vol,
      motion: s.motion,
      scores: s.scores,
      mid: s.mid,
      ci: s.ci,
      msgs: s.msgs,
      sigs: s.sigs,
      risk: s.risk,
      log: s.log,
      gained: s.gained,
      elapsed: s.elapsed,
      dec: s.dec,
      sel: s.sel,
      fb: s.fb,
      result: s.result,
    }
    try {
      localStorage.setItem(_LS, JSON.stringify({ saved }))
    } catch (e) {}
  }
  set(patch, persist) {
    this.setState(patch, () => {
      if (persist !== false) this.persist()
    })
  }

  // ---------- helpers ----------
  px(name, size, style, mono) {
    const def = this.IC[name]
    if (!def) return null
    const rows = def[0].split('/'),
      pal = def[1]
    const w = rows.reduce((a, r) => Math.max(a, r.length), 0),
      h = rows.length
    const rects = []
    rows.forEach((row, y) => {
      for (let x = 0; x < row.length; x++) {
        const c = pal[row[x]]
        if (c)
          rects.push(
            React.createElement('rect', {
              key: x + '_' + y,
              x,
              y,
              width: 1.02,
              height: 1.02,
              fill: mono || c,
            })
          )
      }
    })
    return React.createElement(
      'svg',
      {
        viewBox: '0 0 ' + w + ' ' + h,
        width: size,
        height: size,
        'aria-hidden': 'true',
        style: Object.assign(
          { imageRendering: 'pixelated', display: 'block', flex: 'none' },
          style || {}
        ),
      },
      rects
    )
  }
  city() {
    const bars = []
    const seq = [
      30, 54, 22, 70, 40, 86, 34, 60, 26, 74, 44, 90, 28, 64, 38, 80, 24, 56, 46, 68, 32, 78, 42,
      58,
    ]
    seq.forEach((hh, i) =>
      bars.push(
        React.createElement('rect', {
          key: i,
          x: i * 42,
          y: 120 - hh,
          width: 36,
          height: hh,
          fill: i % 3 === 0 ? '#0F2540' : '#10293F',
        })
      )
    )
    seq.forEach((hh, i) => {
      if (i % 2 === 0)
        bars.push(
          React.createElement('rect', {
            key: 'w' + i,
            x: i * 42 + 10,
            y: 120 - hh + 12,
            width: 6,
            height: 6,
            fill: '#1D4E68',
          })
        )
    })
    return React.createElement(
      'svg',
      {
        viewBox: '0 0 1008 120',
        preserveAspectRatio: 'none',
        width: '100%',
        height: '120',
        'aria-hidden': 'true',
        style: { display: 'block' },
      },
      bars
    )
  }
  beep(kind) {
    if (!this.state.sound) return
    try {
      if (!this.ac) this.ac = new (window.AudioContext || window.webkitAudioContext)()
      const ac = this.ac,
        v = this.state.vol
      const notes = {
        click: [[520, 0.06]],
        msg: [[380, 0.05]],
        safe: [
          [520, 0.08],
          [760, 0.12],
        ],
        warn: [
          [430, 0.09],
          [430, 0.09],
        ],
        risk: [
          [220, 0.14],
          [160, 0.16],
        ],
        xp: [
          [600, 0.06],
          [780, 0.06],
          [980, 0.1],
        ],
        start: [
          [400, 0.07],
          [600, 0.07],
          [800, 0.1],
        ],
        done: [
          [520, 0.09],
          [660, 0.09],
          [790, 0.09],
          [990, 0.18],
        ],
        toggle: [[660, 0.06]],
        err: [[180, 0.18]],
      }[kind] || [[500, 0.05]]
      let t = ac.currentTime
      notes.forEach((n) => {
        const o = ac.createOscillator(),
          g = ac.createGain()
        o.type = 'square'
        o.frequency.value = n[0]
        g.gain.setValueAtTime(v * 0.5, t)
        g.gain.exponentialRampToValueAtTime(0.0001, t + n[1])
        o.connect(g)
        g.connect(ac.destination)
        o.start(t)
        o.stop(t + n[1])
        t += n[1]
      })
    } catch (e) {}
  }
  toast(text, tone) {
    if (this.toastT) clearTimeout(this.toastT)
    this.setState({ toast: { text, tone: tone || 'ok' } })
    this.toastT = setTimeout(() => this.setState({ toast: null }), 3800)
  }
  go(route) {
    this.beep('click')
    this.setState({ route, profile: false, modal: null })
    try {
      location.hash = '/' + route
    } catch (e) {}
    try {
      window.scrollTo(0, 0)
    } catch (e) {}
  }
  find(id) {
    return this.MISSIONS.filter((m) => m.id === id)[0]
  }
  keyAct(fn) {
    return (e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault()
        fn()
      }
    }
  }
  mmss(s) {
    const m = Math.floor(s / 60),
      r = s % 60
    return (m < 10 ? '0' : '') + m + ':' + (r < 10 ? '0' : '') + r
  }
  lvl() {
    return Math.max(1, Math.floor(this.state.xp / 1000) + 7)
  }
  statusOf(m) {
    if (this.state.completed.indexOf(m.id) >= 0) return 'completed'
    if (this.state.mid === m.id && !this.state.result) return 'progress'
    if (m.lock) return 'locked'
    return 'new'
  }

  // ---------- mission flow ----------
  openMission(id) {
    this.beep('click')
    const m = this.find(id)
    if (!m) return
    const st = this.statusOf(m)
    if (st === 'locked') {
      this.setState({ modal: { kind: 'locked', m } })
      this.beep('err')
      return
    }
    this.setState({ modal: { kind: 'detail', m } })
  }
  startMission(id) {
    const m = this.find(id)
    if (!m) return
    const st = this.statusOf(m)
    if (st === 'locked') {
      this.beep('err')
      this.setState({ modal: { kind: 'locked', m } })
      return
    }
    if (this.state.busy) return
    this.setState({ busy: id })
    setTimeout(() => {
      const sid = this.SCRIPTS[id] ? id : m.role === 'SELLER' ? 'fake-shipping' : 'too-good'
      if (sid !== id)
        this.toast('“' + m.t + '” uses a sample conversation in this prototype.', 'warn')
      this.beep('start')
      this.setState(
        {
          busy: null,
          modal: null,
          route: 'play',
          mid: sid,
          srcId: id,
          ci: 0,
          msgs: [],
          typing: false,
          dec: null,
          sel: null,
          fb: null,
          sigs: [],
          risk: 'low',
          log: [],
          gained: 0,
          elapsed: 0,
          result: null,
        },
        () => {
          this.persist()
          this.busyAdv = false
          this.advance()
        }
      )
      try {
        location.hash = '/play'
      } catch (e) {}
    }, 620)
  }
  advance() {
    const sc = this.SCRIPTS[this.state.mid]
    if (!sc) return
    const i = this.state.ci
    if (i >= sc.steps.length) return this.finish()
    const n = sc.steps[i]
    if (n.d) {
      const sigs = n.sigs ? this.state.sigs.concat(n.sigs) : this.state.sigs
      this.setState({ dec: n, sigs, risk: n.risk || this.state.risk }, () => this.persist())
      return
    }
    this.setState({ typing: n.who !== 'you' })
    setTimeout(
      () => {
        this.setState(
          (s) => ({
            msgs: s.msgs.concat([{ who: n.who, t: n.t, time: n.time }]),
            typing: false,
            ci: s.ci + 1,
          }),
          () => {
            this.persist()
            setTimeout(() => this.advance(), 300)
          }
        )
        if (n.who !== 'you') this.beep('msg')
      },
      n.who === 'you' ? 420 : 950
    )
  }
  choose(i) {
    if (this.state.sel !== null || !this.state.dec) return
    const dec = this.state.dec,
      opt = dec.opts[i]
    this.setState({ sel: i })
    this.beep('click')
    setTimeout(() => {
      const order = { low: 0, medium: 1, high: 2, critical: 3 },
        names = ['low', 'medium', 'high', 'critical']
      let risk = this.state.risk
      if (opt.k === 'risk') risk = names[Math.min(3, order[risk] + 1)]
      const xp = opt.k === 'safe' ? 180 : opt.k === 'warn' ? 120 : 40
      const pts = opt.k === 'safe' ? 1150 : opt.k === 'warn' ? 700 : 200
      this.setState(
        (s) => ({
          msgs: s.msgs.concat([{ who: 'you', t: opt.t, time: this.clockNow() }]),
          fb: { k: opt.k, skill: dec.skill, fb: opt.fb },
          risk,
          gained: s.gained + xp,
          log: s.log.concat([
            {
              time: this.mmss(s.elapsed),
              title: opt.log,
              note: opt.note,
              xp,
              k: opt.k,
              pts,
              skill: dec.skill,
            },
          ]),
        }),
        () => this.persist()
      )
      this.beep(opt.k === 'safe' ? 'safe' : opt.k === 'warn' ? 'warn' : 'risk')
    }, 560)
  }
  clockNow() {
    const base = 10 * 60 + 32 + Math.floor(this.state.elapsed / 12)
    const h = Math.floor(base / 60),
      m = base % 60
    return (h > 12 ? h - 12 : h) + ':' + (m < 10 ? '0' : '') + m + (h < 12 ? ' AM' : ' PM')
  }
  next() {
    this.beep('click')
    this.setState(
      (s) => ({ fb: null, dec: null, sel: null, ci: s.ci + 1 }),
      () => {
        this.persist()
        this.advance()
      }
    )
  }
  finish() {
    const log = this.state.log
    const safe = log.filter((l) => l.k === 'safe').length
    const risky = log.filter((l) => l.k === 'risk').length
    const score = 5000 + log.reduce((a, l) => a + l.pts, 0)
    const acc = Math.round(
      (log.reduce((a, l) => a + (l.k === 'safe' ? 1 : l.k === 'warn' ? 0.6 : 0.15), 0) /
        Math.max(1, log.length)) *
        100
    )
    const ending = risky === 0 ? 'safe' : safe >= 2 ? 'mixed' : 'risky'
    const src = this.state.srcId || this.state.mid
    const completed =
      this.state.completed.indexOf(src) >= 0
        ? this.state.completed
        : this.state.completed.concat([src])
    const result = {
      score,
      xp: this.state.gained,
      acc,
      ending,
      safe,
      risky,
      time: this.mmss(this.state.elapsed),
      mid: this.state.mid,
      src,
    }
    const scores = Object.assign({}, this.state.scores)
    scores[src] = Math.max(scores[src] || 0, acc)
    this.beep('done')
    this.animateScore(score)
    this.setState(
      {
        result,
        route: 'result',
        xp: this.state.xp + this.state.gained,
        completed,
        scores,
        badgeCount: this.state.badgeCount + (ending === 'safe' ? 1 : 0),
      },
      () => this.persist()
    )
    try {
      location.hash = '/result'
    } catch (e) {}
    this.toast('Progress saved — +' + this.state.gained + ' XP added.', 'ok')
  }
  exitMission() {
    this.setState({ modal: { kind: 'exit' } })
  }

  // ---------- filtering ----------
  filtered() {
    const s = this.state
    let list = this.MISSIONS.slice()
    if (s.fRole !== 'ALL') list = list.filter((m) => m.role === s.fRole || m.role === 'BOTH')
    if (s.fDif.length) list = list.filter((m) => s.fDif.indexOf(m.dif) >= 0)
    if (s.fCat) list = list.filter((m) => m.cat === s.fCat)
    if (s.fProg !== 'all') list = list.filter((m) => this.statusOf(m) === s.fProg)
    if (s.q.trim()) {
      const q = s.q.trim().toLowerCase()
      list = list.filter(
        (m) => (m.t + ' ' + m.d + ' ' + this.CAT[m.cat]).toLowerCase().indexOf(q) >= 0
      )
    }
    const dif = { EASY: 1, MEDIUM: 2, HARD: 3 }
    if (s.sort === 'dif') list.sort((a, b) => dif[a.dif] - dif[b.dif])
    if (s.sort === 'dur') list.sort((a, b) => a.mins - b.mins)
    if (s.sort === 'xp') list.sort((a, b) => b.xp - a.xp)
    if (s.sort === 'new') list.sort((a, b) => (b.new || 0) - (a.new || 0))
    if (s.sort === 'rec')
      list.sort((a, b) => (b.feat || 0) - (a.feat || 0) || dif[a.dif] - dif[b.dif])
    return list
  }
  isFiltering() {
    const s = this.state
    return !!(s.q.trim() || s.fRole !== 'ALL' || s.fDif.length || s.fCat || s.fProg !== 'all')
  }

  // ---------- view models ----------
  difStyle(d) {
    return d === 'EASY'
      ? { bg: '#07231A', bd: '#1E6B58', fg: '#2FE0A0' }
      : d === 'MEDIUM'
        ? { bg: '#1F1706', bd: '#6B5216', fg: '#F5C84A' }
        : { bg: '#230C0C', bd: '#6B2029', fg: '#F0524B' }
  }
  roleStyle(r) {
    return r === 'BUYER'
      ? { bg: '#07231A', bd: '#1E6B58', fg: '#2FE0A0', ic: 'cart' }
      : r === 'SELLER'
        ? { bg: '#230F0C', bd: '#6B2A20', fg: '#F0857B', ic: 'store' }
        : { bg: '#170F26', bd: '#3C2A63', fg: '#C7A8FF', ic: 'chat' }
  }
  cardVM(m, size) {
    const st = this.statusOf(m),
      d = this.difStyle(m.dif),
      r = this.roleStyle(m.role)
    const cta =
      st === 'completed'
        ? 'PLAY AGAIN'
        : st === 'progress'
          ? 'CONTINUE'
          : st === 'locked'
            ? 'LOCKED'
            : 'START MISSION'
    const btn =
      st === 'locked'
        ? { bg: '#101E2E', bd: '#2A5175', fg: '#5E7794', sh: '#071019' }
        : m.dif === 'EASY'
          ? { bg: '#2FE0A0', bd: '#7CFFD0', fg: '#052C20', sh: '#0B5B42' }
          : m.dif === 'MEDIUM'
            ? { bg: '#F5B942', bd: '#FFD98A', fg: '#2A1C02', sh: '#8C6510' }
            : { bg: '#F0524B', bd: '#FF9088', fg: '#2A0605', sh: '#8C1A18' }
    const bdMap = { EASY: '#1E6B58', MEDIUM: '#6B5216', HARD: '#6B2029' }
    return {
      id: m.id,
      t: m.t,
      d: m.d,
      dif: m.dif,
      role: m.role,
      xp: m.xp,
      mins: m.mins,
      icon: this.px(m.ic, size === 'lg' ? 40 : size === 'sm' ? 30 : 34),
      star: this.px('star', 15),
      clock: this.px('clock', 14),
      roleIcon: this.px(r.ic, 13),
      difBg: d.bg,
      difBd: d.bd,
      difFg: d.fg,
      roleBg: r.bg,
      roleBd: r.bd,
      roleFg: r.fg,
      catLabel: this.CAT[m.cat],
      isNew: !!m.new,
      op: st === 'locked' ? '.62' : '1',
      bd:
        size === 'lg'
          ? bdMap[m.dif]
          : st === 'completed'
            ? '#1E6B58'
            : st === 'locked'
              ? '#1B3550'
              : '#1E4670',
      bd2: bdMap[m.dif],
      bg: st === 'completed' ? '#07231A' : size === 'lg' ? '#0A1728' : '#0A1728',
      glow: size === 'lg' ? '0 0 0 1px rgba(0,0,0,.3)' : 'none',
      titleFg: st === 'locked' ? '#8FA3BD' : '#EAF4FF',
      stFg: st === 'completed' ? '#2FE0A0' : '#7A93AE',
      stLabel:
        st === 'completed'
          ? 'COMPLETED'
          : st === 'locked'
            ? 'LOCKED'
            : st === 'progress'
              ? 'IN PROGRESS'
              : 'NOT STARTED',
      stIcon:
        st === 'completed'
          ? this.px('check', 14)
          : st === 'locked'
            ? this.px('lock', 14)
            : this.px('flag', 14),
      btnBg: btn.bg,
      btnBd: btn.bd,
      btnFg: btn.fg,
      btnSh: btn.sh,
      ctaLabel: this.state.busy === m.id ? 'STARTING…' : cta,
      busy: this.state.busy === m.id,
      open: () => this.openMission(m.id),
      key: this.keyAct(() => this.openMission(m.id)),
      key2: this.keyAct(() => (st === 'locked' ? this.openMission(m.id) : this.startMission(m.id))),
      cta: () => (st === 'locked' ? this.openMission(m.id) : this.startMission(m.id)),
    }
  }

  modalVM() {
    const md = this.state.modal
    if (!md) return {}
    const close = () => {
      this.beep('click')
      this.setState({ modal: null })
    }
    if (md.kind === 'detail') {
      const m = md.m,
        st = this.statusOf(m),
        d = this.difStyle(m.dif),
        best = this.state.scores[m.id]
      const rows = [
        ['ROLE', m.role],
        ['DIFFICULTY', m.dif],
        ['DURATION', m.mins + ' min'],
        ['REWARD', '+' + m.xp + ' XP'],
        ['SCAM TYPE', this.CAT[m.cat]],
        [
          'STATUS',
          st === 'completed' ? 'Completed' : st === 'progress' ? 'In progress' : 'Not started',
        ],
      ]
      const body = React.createElement(
        'div',
        { style: { display: 'flex', flexDirection: 'column', gap: '14px' } },
        React.createElement(
          'p',
          { key: 'd', style: { margin: 0, fontSize: '19px', lineHeight: 1.45, color: '#A8BFD6' } },
          m.d +
            ' You will read a realistic conversation and decide how to respond at every risky moment.'
        ),
        React.createElement(
          'div',
          {
            key: 'g',
            style: {
              display: 'grid',
              gridTemplateColumns: '1fr 1fr',
              gap: '2px',
              background: '#1B3550',
              border: '2px solid #1B3550',
            },
          },
          rows.map((r, i) =>
            React.createElement(
              'div',
              { key: i, style: { background: '#0C1826', padding: '9px 11px' } },
              React.createElement(
                'span',
                {
                  style: {
                    display: 'block',
                    fontSize: '16px',
                    letterSpacing: '.1em',
                    color: '#5E7794',
                  },
                },
                r[0]
              ),
              React.createElement(
                'span',
                {
                  style: {
                    display: 'block',
                    fontFamily: "'Pixelify Sans',monospace",
                    fontSize: '20px',
                    color: '#EAF4FF',
                  },
                },
                r[1]
              )
            )
          )
        ),
        React.createElement(
          'div',
          { key: 's' },
          React.createElement(
            'span',
            {
              style: {
                display: 'block',
                fontSize: '17px',
                letterSpacing: '.1em',
                color: '#5E7794',
                marginBottom: '6px',
              },
            },
            'YOU WILL TRAIN'
          ),
          React.createElement(
            'div',
            { style: { display: 'flex', flexWrap: 'wrap', gap: '8px' } },
            ['Fraud Instinct', 'Link Safety', 'Pressure Resistance', 'Platform Safety'].map((s) =>
              React.createElement(
                'span',
                {
                  key: s,
                  style: {
                    padding: '3px 10px',
                    border: '2px solid #1E4670',
                    color: '#9FB6D0',
                    fontSize: '17px',
                  },
                },
                s
              )
            )
          )
        ),
        best
          ? React.createElement(
              'div',
              {
                key: 'b',
                style: {
                  padding: '10px 12px',
                  background: '#07231A',
                  border: '2px solid #1E6B58',
                  fontSize: '19px',
                  color: '#2FE0A0',
                },
              },
              'Best result: ' + best + '% accuracy'
            )
          : null
      )
      return {
        open: 1,
        w: '560px',
        bd: d.bd,
        title: m.t,
        icon: this.px(m.ic, 30),
        body,
        actions: [
          {
            label: 'BACK TO MISSIONS',
            bg: '#0C1826',
            bd: '#2A5175',
            fg: '#BBD2EA',
            sh: '#071019',
            on: close,
          },
          {
            label: st === 'completed' ? 'PLAY AGAIN' : 'START MISSION',
            bg: '#2FE0A0',
            bd: '#7CFFD0',
            fg: '#052C20',
            sh: '#0B5B42',
            on: () => this.startMission(m.id),
          },
        ],
      }
    }
    if (md.kind === 'locked') {
      const m = md.m
      return {
        open: 1,
        w: '460px',
        bd: '#2A5175',
        title: 'Mission locked',
        icon: this.px('lock', 26),
        body: React.createElement(
          'p',
          { style: { margin: 0, fontSize: '19px', lineHeight: 1.45, color: '#A8BFD6' } },
          m.lock + ' Keep completing missions — your progress unlocks it automatically.'
        ),
        actions: [
          {
            label: 'GOT IT',
            bg: '#0C1826',
            bd: '#2A5175',
            fg: '#BBD2EA',
            sh: '#071019',
            on: close,
          },
          {
            label: 'BROWSE MISSIONS',
            bg: '#3D8BFD',
            bd: '#8FC9FF',
            fg: '#04121F',
            sh: '#12417F',
            on: () => this.go('missions'),
          },
        ],
      }
    }
    if (md.kind === 'help') {
      return {
        open: 1,
        w: '460px',
        bd: '#1E4670',
        title: md.title,
        icon: this.px('bulb', 24),
        body: React.createElement(
          'p',
          { style: { margin: 0, fontSize: '19px', lineHeight: 1.45, color: '#A8BFD6' } },
          md.text
        ),
        actions: [
          { label: 'CLOSE', bg: '#0C1826', bd: '#2A5175', fg: '#BBD2EA', sh: '#071019', on: close },
        ],
      }
    }
    if (md.kind === 'exit') {
      return {
        open: 1,
        w: '460px',
        bd: '#6B2029',
        title: 'Leave this mission?',
        icon: this.px('warn', 24),
        body: React.createElement(
          'p',
          { style: { margin: 0, fontSize: '19px', lineHeight: 1.45, color: '#A8BFD6' } },
          'Your attempt is saved locally, so you can come back to this exact step later. Leave the conversation now?'
        ),
        actions: [
          { label: 'STAY', bg: '#0C1826', bd: '#2A5175', fg: '#BBD2EA', sh: '#071019', on: close },
          {
            label: 'LEAVE MISSION',
            bg: '#F0524B',
            bd: '#FF9088',
            fg: '#2A0605',
            sh: '#8C1A18',
            on: () => {
              this.setState({ modal: null })
              this.go('missions')
              this.toast('Attempt saved — resume it any time.', 'ok')
            },
          },
        ],
      }
    }
    if (md.kind === 'reset') {
      return {
        open: 1,
        w: '460px',
        bd: '#6B2029',
        title: 'Reset demo data?',
        icon: this.px('warn', 24),
        body: React.createElement(
          'p',
          { style: { margin: 0, fontSize: '19px', lineHeight: 1.45, color: '#A8BFD6' } },
          'This clears XP, badges, streak and saved attempts stored in this browser only. Nothing is sent anywhere and no account is affected.'
        ),
        actions: [
          {
            label: 'CANCEL',
            bg: '#0C1826',
            bd: '#2A5175',
            fg: '#BBD2EA',
            sh: '#071019',
            on: close,
          },
          {
            label: 'RESET DEMO',
            bg: '#F0524B',
            bd: '#FF9088',
            fg: '#2A0605',
            sh: '#8C1A18',
            on: () => this.resetDemo(),
          },
        ],
      }
    }
    if (md.kind === 'badge') {
      const b = md.b
      return {
        open: 1,
        w: '440px',
        bd: b.earned ? '#6B5216' : '#2A5175',
        title: b.name,
        icon: this.px(b.ic, 28),
        body: React.createElement(
          'div',
          { style: { display: 'flex', flexDirection: 'column', gap: '12px' } },
          React.createElement(
            'p',
            {
              key: 'p',
              style: { margin: 0, fontSize: '19px', lineHeight: 1.45, color: '#A8BFD6' },
            },
            b.desc
          ),
          React.createElement(
            'div',
            {
              key: 'u',
              style: { padding: '10px 12px', background: '#0C1826', border: '2px solid #1B3550' },
            },
            React.createElement(
              'span',
              {
                style: {
                  display: 'block',
                  fontSize: '16px',
                  letterSpacing: '.1em',
                  color: '#5E7794',
                },
              },
              'UNLOCK CONDITION'
            ),
            React.createElement(
              'span',
              { style: { display: 'block', fontSize: '19px', color: '#EAF4FF' } },
              b.cond
            )
          ),
          React.createElement(
            'div',
            { key: 'b2' },
            React.createElement(
              'div',
              {
                style: {
                  display: 'flex',
                  justifyContent: 'space-between',
                  fontSize: '18px',
                  color: '#8FA3BD',
                },
              },
              React.createElement('span', null, b.earned ? 'Earned ' + b.date : 'Progress'),
              React.createElement('span', null, b.have + ' / ' + b.need)
            ),
            React.createElement(
              'div',
              {
                style: {
                  marginTop: '6px',
                  height: '14px',
                  background: '#08151F',
                  border: '2px solid #1B3550',
                },
              },
              React.createElement('div', {
                style: {
                  height: '100%',
                  width: Math.min(100, (b.have / b.need) * 100) + '%',
                  background: b.earned ? '#F5C84A' : '#3D8BFD',
                },
              })
            )
          )
        ),
        actions: [
          { label: 'CLOSE', bg: '#0C1826', bd: '#2A5175', fg: '#BBD2EA', sh: '#071019', on: close },
        ],
      }
    }
    if (md.kind === 'signal') {
      return {
        open: 1,
        w: '480px',
        bd: '#6B2029',
        title: 'Why this is a scam signal',
        icon: this.px('flagr', 24),
        body: React.createElement(
          'div',
          { style: { display: 'flex', flexDirection: 'column', gap: '10px' } },
          React.createElement(
            'p',
            {
              key: 't',
              style: {
                margin: 0,
                fontFamily: "'Pixelify Sans',monospace",
                fontSize: '21px',
                color: '#EAF4FF',
              },
            },
            md.sig.t
          ),
          React.createElement(
            'p',
            {
              key: 'x',
              style: { margin: 0, fontSize: '19px', lineHeight: 1.45, color: '#A8BFD6' },
            },
            md.sig.why ||
              'This behaviour appears in most marketplace fraud: it moves the deal away from the parts of the platform that protect you — checkout, delivery, and support. Treat it as a reason to slow down and verify inside the app.'
          )
        ),
        actions: [
          { label: 'CLOSE', bg: '#0C1826', bd: '#2A5175', fg: '#BBD2EA', sh: '#071019', on: close },
        ],
      }
    }
    return {}
  }
  resetDemo() {
    try {
      localStorage.removeItem(_LS)
    } catch (e) {}
    this.setState({
      xp: 0,
      streak: 0,
      completed: [],
      badgeCount: 0,
      scores: {},
      mid: null,
      ci: 0,
      msgs: [],
      dec: null,
      sel: null,
      fb: null,
      sigs: [],
      risk: 'low',
      log: [],
      gained: 0,
      elapsed: 0,
      result: null,
      modal: null,
      route: 'home',
    })
    this.toast('Demo data cleared. Starting from zero.', 'ok')
    this.beep('toggle')
  }

  BADGES = [
    {
      ic: 'shield',
      name: 'First Shield',
      desc: 'Awarded for finishing your first mission without a single risky choice.',
      cond: 'Finish 1 mission with a safe ending',
      need: 1,
      have: 1,
      earned: 1,
      date: '12 Mar',
    },
    {
      ic: 'glass',
      name: 'Keen Eye',
      desc: 'For spotting scam signals before they turn into losses.',
      cond: 'Detect 25 scam signals',
      need: 25,
      have: 25,
      earned: 1,
      date: '18 Mar',
    },
    {
      ic: 'trophy',
      name: 'Mission Veteran',
      desc: 'Awarded for consistent training across both roles.',
      cond: 'Complete 15 missions',
      need: 15,
      have: 18,
      earned: 1,
      date: '02 Apr',
    },
    {
      ic: 'shieldw',
      name: 'Link Guardian',
      desc: 'For refusing phishing links and checking domains before clicking.',
      cond: 'Refuse 10 unsafe links',
      need: 10,
      have: 10,
      earned: 1,
      date: '11 Apr',
    },
    {
      ic: 'fire',
      name: 'Streak Keeper',
      desc: 'Train on consecutive days to keep this one alive.',
      cond: 'Reach a 7-day streak',
      need: 7,
      have: 5,
      earned: 0,
    },
    {
      ic: 'medal',
      name: 'Cool Head',
      desc: 'For resisting urgency and pressure tactics.',
      cond: 'Resist pressure 20 times',
      need: 20,
      have: 12,
      earned: 0,
    },
    {
      ic: 'lock',
      name: 'Account Warden',
      desc: 'For protecting codes, passwords and account access.',
      cond: 'Complete all account missions',
      need: 4,
      have: 1,
      earned: 0,
    },
    {
      ic: 'star',
      name: 'Anti-Scam Legend',
      desc: 'The final badge. Reserved for a perfect record across every scam type.',
      cond: 'Reach Level 10 with 90% accuracy',
      need: 10,
      have: 7,
      earned: 0,
    },
  ]

  SKILLS = [
    {
      k: 'Fraud Instinct',
      ic: 'glass',
      c: '#2FE0A0',
      bd: '#1E6B58',
      lvl: 7,
      pct: 78,
      d: 'You spotted the red flags before it was too late.',
    },
    {
      k: 'Link Safety',
      ic: 'link',
      c: '#3D8BFD',
      bd: '#1E4670',
      lvl: 6,
      pct: 62,
      d: 'You checked before you clicked.',
    },
    {
      k: 'Pressure Resistance',
      ic: 'fire',
      c: '#A97BFF',
      bd: '#3C2A63',
      lvl: 7,
      pct: 70,
      d: 'You didn’t get rushed into a bad decision.',
    },
    {
      k: 'Platform Safety',
      ic: 'shield',
      c: '#F5C84A',
      bd: '#6B5216',
      lvl: 6,
      pct: 66,
      d: 'You kept your account and info secure.',
    },
    {
      k: 'Payment Safety',
      ic: 'card',
      c: '#4FD8E8',
      bd: '#1E4670',
      lvl: 6,
      pct: 58,
      d: 'You kept the money inside protected checkout.',
    },
    {
      k: 'Account Safety',
      ic: 'lock',
      c: '#F0857B',
      bd: '#6B2029',
      lvl: 5,
      pct: 48,
      d: 'You never shared codes or credentials.',
    },
  ]

  RULES = [
    {
      cat: 'Payment',
      ic: 'card',
      t: 'Never pay through a link sent in chat',
      why: 'Look-alike checkout pages are the cheapest part of a scam to build. The address bar is the only thing that tells them apart.',
      alt: 'Open the app yourself and pay from the order screen.',
      mid: 'payment-outside',
    },
    {
      cat: 'Payment',
      ic: 'qr',
      t: 'Check the recipient name before confirming a transfer',
      why: 'QR codes and payment links can carry a different recipient than the one you agreed with.',
      alt: 'Read the name and amount on the confirmation screen out loud before you tap.',
      mid: 'qr-trap',
    },
    {
      cat: 'Codes',
      ic: 'lock',
      t: 'Never share codes from SMS',
      why: 'One-time codes confirm logins, password resets and payouts. Sharing one hands over your account.',
      alt: 'End the conversation and check the order inside the app.',
      mid: 'account-takeover',
    },
    {
      cat: 'Links',
      ic: 'link',
      t: 'Read the domain before you click',
      why: 'Extra words like “secure”, “pay” or a hyphenated brand name are the usual disguise for a phishing page.',
      alt: 'Type the marketplace address yourself, or use the app.',
      mid: 'phishing-links',
    },
    {
      cat: 'Delivery',
      ic: 'truck',
      t: 'Arrange delivery through the platform',
      why: 'Off-platform couriers break the link between the order, the tracking and your protection.',
      alt: 'Book pickup from the order screen so tracking stays attached.',
      mid: 'fake-shipping',
    },
    {
      cat: 'Chat',
      ic: 'chat',
      t: 'Keep the deal in the marketplace chat',
      why: 'Moving to another messenger removes moderation, history and any evidence for a dispute.',
      alt: 'Say you are happy to continue here, and stay put.',
      mid: 'chat-red-flags',
    },
    {
      cat: 'Support',
      ic: 'shield',
      t: 'Contact support only from inside the app',
      why: 'Impersonators reach out first and sound official. Real support never opens a chat asking for codes.',
      alt: 'Close the message and open Help from the app menu.',
      mid: 'fake-support',
    },
    {
      cat: 'Returns',
      ic: 'box',
      t: 'Document the item before it ships',
      why: 'Return fraud relies on there being no proof of what left your hands.',
      alt: 'Photograph the item and packaging, and keep the receipt.',
      mid: 'damaged-claim',
    },
  ]

  pageVals() {
    const s = this.state,
      px = (n, sz, st) => this.px(n, sz, st)
    const chipS = (a) =>
      a
        ? { bg: '#07231A', bd: '#2FE0A0', fg: '#2FE0A0' }
        : { bg: '#0C1826', bd: '#1B3550', fg: '#9FB6D0' }
    const cat = {}
    Object.keys(this.CAT).forEach((k) => (cat[k] = this.MISSIONS.filter((m) => m.cat === k).length))
    const doneByCat = {}
    Object.keys(this.CAT).forEach(
      (k) =>
        (doneByCat[k] = this.MISSIONS.filter(
          (m) => m.cat === k && s.completed.indexOf(m.id) >= 0
        ).length)
    )
    const maxCat = Math.max(1, ...Object.keys(cat).map((k) => cat[k]))
    const rq = (s.ruleQ || '').trim().toLowerCase()
    const rules = this.RULES.filter(
      (r) =>
        (s.ruleCat === 'all' || r.cat === s.ruleCat) &&
        (!rq || (r.t + ' ' + r.why + ' ' + r.alt).toLowerCase().indexOf(rq) >= 0)
    )
    const acts = s.log.slice(-4).reverse()
    const scen = ['too-good', 'fake-shipping'].map((id) => {
      const sc = this.SCRIPTS[id],
        m = this.find(id),
        d = this.difStyle(m.dif),
        r = this.roleStyle(sc.role)
      const other = sc.with === 'SELLER' ? 'seller' : 'buyer'
      const btn =
        m.dif === 'EASY'
          ? { bg: '#2FE0A0', bd: '#7CFFD0', fg: '#052C20', sh: '#0B5B42' }
          : { bg: '#F5B942', bd: '#FFD98A', fg: '#2A1C02', sh: '#8C6510' }
      return {
        t: m.t,
        d:
          'A ' +
          sc.steps.length +
          '-step conversation with ' +
          (sc.with === 'SELLER' ? 'a seller' : 'a buyer') +
          ', three decisions, and a different ending depending on how you answer. ' +
          m.d,
        role: sc.role,
        dif: m.dif,
        xp: m.xp,
        bd: d.bd,
        difFg: d.fg,
        roleFg: r.fg,
        icon: px(m.ic, 40),
        first: sc.steps[0].t,
        avatar: px(other, 26),
        avBg: sc.with === 'SELLER' ? '#230F0C' : '#07231A',
        avBd: sc.with === 'SELLER' ? '#6B2A20' : '#1E6B58',
        obj: sc.obj.map((o) => ({ t: o, icon: px('check', 14) })),
        cta: s.completed.indexOf(id) >= 0 ? 'PLAY AGAIN' : 'START SCENARIO',
        btnBg: btn.bg,
        btnBd: btn.bd,
        btnFg: btn.fg,
        btnSh: btn.sh,
        start: () => this.startMission(id),
      }
    })
    return {
      barsBig: px('bars', 30),
      bookBig: px('book', 30),
      gearBig: px('gear', 30),
      nfIcon: px('skull', 56),
      errIcon: px('skull', 46),
      isNotFound: s.route === 'nf',
      uiFilter: s.contrast === 'high' ? 'contrast(1.14) saturate(1.06)' : 'none',
      uiZoom: s.textSize === 'large' ? '1.12' : '1',
      scenarios: scen,
      statGrid:
        (s.vw || 1600) >= 1100
          ? 'minmax(280px,1.4fr) repeat(4,minmax(0,1fr))'
          : (s.vw || 1600) >= 700
            ? '1fr 1fr'
            : '1fr',
      xpToNext: 1000 - (s.xp % 1000),
      progStats: [
        { label: 'COMPLETED', v: s.completed.length, fg: '#EAF4FF', icon: px('shield', 22) },
        {
          label: 'AVG ACCURACY',
          v:
            (Object.keys(s.scores).length
              ? Math.round(
                  Object.keys(s.scores).reduce((a, k) => a + s.scores[k], 0) /
                    Object.keys(s.scores).length
                )
              : 0) + '%',
          fg: '#2FE0A0',
          icon: px('glass', 22),
        },
        {
          label: 'SAFE ENDINGS',
          v: s.badgeCount ? Math.max(1, Math.round(s.completed.length * 0.8)) : 0,
          fg: '#4FD8E8',
          icon: px('check', 22),
        },
        { label: 'DAY STREAK', v: s.streak, fg: '#F0853A', icon: px('fire', 22) },
      ],
      skillBars: this.SKILLS.map((k) => ({
        name: k.k,
        lvl: k.lvl,
        pct: k.pct,
        w: k.pct + '%',
        fg: k.c,
        icon: px(k.ic, 18),
      })),
      chart: Object.keys(this.CAT).map((k) => ({
        label: this.CAT[k],
        v: doneByCat[k] + ' / ' + cat[k],
        w: Math.round((cat[k] / maxCat) * 100) + '%',
        fg: {
          pay: '#3D8BFD',
          ship: '#F5A03A',
          fake: '#A97BFF',
          acc: '#F0524B',
          off: '#4FD8E8',
          ret: '#2FE0A0',
        }[k],
      })),
      chartSummary:
        'Across ' +
        this.MISSIONS.length +
        ' missions, the largest group is ' +
        this.CAT[Object.keys(cat).sort((a, b) => cat[b] - cat[a])[0]] +
        '. You have completed ' +
        s.completed.length +
        '.',
      buyerCount: this.MISSIONS.filter((m) => m.role === 'BUYER').length,
      sellerCount: this.MISSIONS.filter((m) => m.role === 'SELLER').length,
      badgeSummary:
        this.BADGES.filter((b) => b.earned).length +
        ' earned · ' +
        this.BADGES.filter((b) => !b.earned).length +
        ' in progress',
      badgeCards: this.BADGES.map((b) => ({
        name: b.name,
        icon: px(b.ic, 34),
        op: b.earned ? '1' : '.75',
        bg: b.earned ? '#1A1406' : '#0C1826',
        bd: b.earned ? '#6B5216' : '#1B3550',
        fg: b.earned ? '#F5C84A' : '#9FB6D0',
        barFg: b.earned ? '#F5C84A' : '#3D8BFD',
        w: Math.min(100, Math.round((b.have / b.need) * 100)) + '%',
        status: b.earned ? 'EARNED ' + b.date : b.have + ' / ' + b.need,
        open: () => {
          this.beep('click')
          this.setState({ modal: { kind: 'badge', b } })
        },
      })),
      hasActivity: acts.length > 0,
      noActivity: acts.length === 0,
      activity: acts.map((a) => ({
        t: a.title,
        sub: a.skill + ' · ' + a.time,
        xp: '+' + a.xp + ' XP',
        icon: px(a.k === 'safe' ? 'check' : a.k === 'warn' ? 'warn' : 'skull', 18),
        bd: a.k === 'safe' ? '#1E6B58' : a.k === 'warn' ? '#6B5216' : '#6B2029',
      })),
      ruleQ: s.ruleQ,
      onRuleSearch: (e) => this.setState({ ruleQ: e.target.value }),
      clearRules: () => this.setState({ ruleQ: '', ruleCat: 'all' }),
      ruleCats: ['all']
        .concat(this.RULES.map((r) => r.cat).filter((v, i, a) => a.indexOf(v) === i))
        .map((c) => {
          const a = s.ruleCat === c,
            st = chipS(a)
          return {
            label: c === 'all' ? 'All' : c,
            active: a,
            bg: st.bg,
            bd: st.bd,
            fg: st.fg,
            on: () => {
              this.beep('click')
              this.setState({ ruleCat: c })
            },
          }
        }),
      hasRules: rules.length > 0,
      noRules: rules.length === 0,
      rules: rules.map((r) => {
        const m = this.find(r.mid) || this.MISSIONS[0]
        return {
          cat: r.cat.toUpperCase(),
          t: r.t,
          why: r.why,
          alt: r.alt,
          icon: px(r.ic, 24),
          mission: m.t,
          on: () => this.openMission(m.id),
        }
      }),
      soundSwBg: s.sound ? '#07231A' : '#0C1826',
      soundIcon2: px(s.sound ? 'sound' : 'mute', 18),
      volPct: Math.round(s.vol * 100),
      onVol: (e) => {
        const v = Number(e.target.value) / 100
        this.set({ vol: v })
      },
      previewSound: () => {
        this.beep('xp')
        if (!s.sound) this.toast('Sound is off — switch it on to hear it.', 'warn')
      },
      settingRows: [
        {
          label: 'Motion',
          hint: 'Reduced motion replaces movement with fades and removes confetti.',
          opts: [
            ['standard', 'Standard'],
            ['reduced', 'Reduced'],
          ].map((o) => {
            const a = s.motion === o[0],
              st = chipS(a)
            return {
              label: o[1],
              active: a,
              bg: st.bg,
              bd: st.bd,
              fg: st.fg,
              on: () => {
                this.beep('toggle')
                this.set({ motion: o[0] })
              },
            }
          }),
        },
        {
          label: 'Contrast',
          hint: 'Enhanced contrast deepens borders and text separation.',
          opts: [
            ['normal', 'Normal'],
            ['high', 'Enhanced'],
          ].map((o) => {
            const a = (s.contrast || 'normal') === o[0],
              st = chipS(a)
            return {
              label: o[1],
              active: a,
              bg: st.bg,
              bd: st.bd,
              fg: st.fg,
              on: () => {
                this.beep('toggle')
                this.set({ contrast: o[0] })
              },
            }
          }),
        },
        {
          label: 'Text size',
          hint: 'Scales the whole interface for easier reading.',
          opts: [
            ['normal', 'Normal'],
            ['large', 'Large'],
          ].map((o) => {
            const a = (s.textSize || 'normal') === o[0],
              st = chipS(a)
            return {
              label: o[1],
              active: a,
              bg: st.bg,
              bd: st.bd,
              fg: st.fg,
              on: () => {
                this.beep('toggle')
                this.set({ textSize: o[0] })
              },
            }
          }),
        },
      ],
      demoStates: [
        ['none', 'Normal'],
        ['loading', 'Loading skeletons'],
        ['empty', 'Empty catalogue'],
        ['error', 'Network error'],
      ].map((o) => {
        const a = (s.demo || 'none') === o[0],
          st = chipS(a)
        return {
          label: o[1],
          active: a,
          bg: st.bg,
          bd: st.bd,
          fg: st.fg,
          on: () => {
            this.beep('click')
            this.setState({ demo: o[0] })
            this.toast(
              o[0] === 'none'
                ? 'Mission hub back to normal.'
                : 'Mission hub now previews: ' + o[1] + '.',
              'ok'
            )
          },
        }
      }),
      demoLoading: s.demo === 'loading',
      demoEmpty: s.demo === 'empty',
      demoError: s.demo === 'error',
      skeletons: [1, 2, 3, 4, 5, 6],
      retryDemo: () => {
        this.beep('click')
        this.setState({ demo: 'none' })
        this.toast('Missions reloaded.', 'ok')
      },
      openReset: () => {
        this.beep('click')
        this.setState({ modal: { kind: 'reset' } })
      },
    }
  }

  animateScore(target) {
    clearInterval(this.cnt)
    let v = 0
    const step = Math.max(1, Math.round(target / 26))
    this.setState({ dispScore: 0, resReveal: false })
    setTimeout(() => this.setState({ resReveal: true }), 120)
    this.cnt = setInterval(() => {
      v = Math.min(target, v + step)
      this.setState({ dispScore: v })
      if (v >= target) clearInterval(this.cnt)
    }, 30)
  }

  resultVals() {
    const s = this.state,
      px = (n, sz, st) => this.px(n, sz, st)
    const r = s.result || {
      score: 8450,
      xp: 540,
      acc: 92,
      ending: 'safe',
      safe: 3,
      risky: 0,
      time: '01:49',
      mid: 'too-good',
      src: 'too-good',
    }
    const m = this.find(r.src) || this.MISSIONS[0]
    const sc = this.SCRIPTS[r.mid] || this.SCRIPTS['too-good']
    const safe = r.ending === 'safe',
      reveal = s.resReveal !== false
    const lvlAfter = this.lvl(),
      lvlBefore = Math.max(1, Math.floor(Math.max(0, s.xp - r.xp) / 1000) + 7)
    const log = s.log.length ? s.log : []
    const good = log.filter((l) => l.k === 'safe').map((l) => l.title)
    const bad = log.filter((l) => l.k !== 'safe').map((l) => l.note)
    const genericGood = [
      'You read the whole message before answering',
      'You kept the deal inside the marketplace',
      'You stayed calm under pressure',
    ]
    const genericBad = [
      'Double-check unfamiliar links before opening them',
      'Watch for artificial urgency and deadlines',
      'Verify support and couriers through official screens only',
    ]
    while (good.length < 4) good.push(genericGood[good.length % 3])
    while (bad.length < 3) bad.push(genericBad[bad.length % 3])
    const next = this.MISSIONS.filter(
      (x) => x.id !== r.src && !x.lock && this.state.completed.indexOf(x.id) < 0
    ).slice(0, 3)
    const week = ['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN']
    const doneDays = Math.min(6, s.streak)
    const conf = []
    const cols = ['#2FE0A0', '#F5C84A', '#4FD8E8', '#A97BFF', '#F0857B']
    for (let i = 0; i < 26; i++)
      conf.push({
        l: ((i * 3.9 + (i % 5) * 2) % 100) + '%',
        bg: cols[i % 5],
        dur: 1.2 + (i % 4) * 0.25 + 's',
        delay: (i % 7) * 0.09 + 's',
      })
    return {
      triGrid: (s.vw || 1600) >= 1000 ? 'repeat(3,minmax(0,1fr))' : '1fr',
      resBg: safe ? '#07231A' : '#1A1406',
      resBd: safe ? '#1E6B58' : '#6B5216',
      resFg: safe ? '#2FE0A0' : '#F5C84A',
      resFs: (s.vw || 1600) >= 1000 ? '46px' : '30px',
      resTitle: safe ? 'MISSION COMPLETE!' : 'MISSION REVIEWED',
      resSub: safe
        ? 'You outsmarted the scam. Great instincts!'
        : 'Let’s look at where the risk crept in.',
      resIcon: px(safe ? 'shield' : 'warn', 56),
      resIconAnim: safe && reveal ? 'glowPulse 1.8s ease-out 2' : 'none',
      resChar: px(sc.role === 'BUYER' ? 'buyer' : 'seller', 74),
      confettiOn: safe && s.motion !== 'reduced' && reveal,
      confetti: conf,
      resScore: (s.dispScore == null ? r.score : s.dispScore).toLocaleString('en-US'),
      resXp: r.xp,
      resLvlA: lvlBefore,
      resLvlB: lvlAfter,
      resLvlPct: reveal ? Math.round((s.xp % 1000) / 10) + '%' : '0%',
      resScenIcon: px(m.ic, 30),
      resScenTitle: m.t,
      resScenDesc: m.d,
      resStBg: safe ? '#0B2B22' : '#1F1706',
      resStBd: safe ? '#1E6B58' : '#6B5216',
      resStFg: safe ? '#2FE0A0' : '#F5C84A',
      resStIcon: px(safe ? 'check' : 'warn', 14),
      resStLabel: safe ? 'COMPLETED — SAFE ENDING' : 'COMPLETED — RISKY ENDING',
      trophyRes: px('trophy', 18),
      starRes: px('star', 18),
      fireRes: px('fire', 18),
      boostIcon: px('star', 40),
      resSkills: this.SKILLS.slice(0, 4).map((k, i) => {
        const filled = Math.round(k.pct / 10)
        return {
          name: k.k.toUpperCase(),
          lvl: k.lvl,
          desc: k.d,
          bd: k.bd,
          icon: px(k.ic, 18),
          segs: Array.from({ length: 10 }, (_, j) => ({
            c: reveal && j < filled ? k.c : '#16324C',
            delay: j * 40 + 'ms',
          })),
        }
      }),
      timeline: log.map((l) => ({
        time: l.time,
        title: l.title,
        note: l.note,
        xp: l.xp,
        dot: l.k === 'safe' ? '#2FE0A0' : l.k === 'warn' ? '#F5C84A' : '#F0524B',
      })),
      goodList: good.slice(0, 4).map((t) => ({ t, icon: px('check', 15) })),
      badList: bad.slice(0, 3).map((t) => ({ t, icon: px('warn', 15) })),
      goodNote: safe ? 'That’s how you win against scams.' : 'Hold on to the calls you got right.',
      goodIcon: px('shield', 30),
      badNote: 'Keep building your scam awareness.',
      badIcon: px('skull', 30),
      nextMissions: next.map((x) => {
        const d = this.difStyle(x.dif)
        return {
          t: x.t,
          d: x.d,
          dif: x.dif,
          xp: x.xp,
          difFg: d.fg,
          icon: px(x.ic, 30),
          open: () => this.openMission(x.id),
        }
      }),
      week: week.map((w, i) => ({
        label: w,
        bg: i < doneDays ? '#0B2B22' : i === doneDays ? '#1F1706' : '#0C1826',
        bd: i < doneDays ? '#1E6B58' : i === doneDays ? '#F5C84A' : '#1B3550',
        labelFg: i <= doneDays ? '#DCE9F7' : '#5E7794',
        icon: i < doneDays ? px('check', 14) : i === doneDays ? px('fire', 16) : null,
      })),
      resActions: [
        {
          label: 'NEXT MISSION',
          bg: '#2FE0A0',
          bd: '#7CFFD0',
          fg: '#052C20',
          sh: '#0B5B42',
          on: () => (next[0] ? this.startMission(next[0].id) : this.go('missions')),
        },
        {
          label: 'PLAY AGAIN',
          bg: '#0C1826',
          bd: '#2A5175',
          fg: '#BBD2EA',
          sh: '#071019',
          on: () => this.startMission(r.src),
        },
        {
          label: 'ALL MISSIONS',
          bg: '#0C1826',
          bd: '#2A5175',
          fg: '#BBD2EA',
          sh: '#071019',
          on: () => this.go('missions'),
        },
        {
          label: 'VIEW PROGRESS',
          bg: '#0C1826',
          bd: '#2A5175',
          fg: '#BBD2EA',
          sh: '#071019',
          on: () => this.go('progress'),
        },
      ],
      boostBg: s.boost ? '#1A1406' : '#F5B942',
      boostFg: s.boost ? '#F5C84A' : '#2A1C02',
      boostLabel: s.boost ? 'BOOST ACTIVE' : 'ACTIVATE BOOST',
      activateBoost: () => {
        this.beep('xp')
        this.setState({ boost: !s.boost })
        this.toast(
          s.boost ? 'XP boost switched off.' : 'XP boost active for your next 3 missions.',
          'ok'
        )
      },
    }
  }

  playVals() {
    const s = this.state,
      px = (n, sz, st) => this.px(n, sz, st)
    const sc = this.SCRIPTS[s.mid] || this.SCRIPTS['too-good']
    const m = this.find(s.srcId || s.mid) || this.MISSIONS[0]
    const d = this.difStyle(m.dif)
    const total = sc.steps.length
    const doneDec = s.log.length
    const RISK =
      {
        low: {
          fg: '#2FE0A0',
          bd: '#1E6B58',
          label: 'LOW RISK',
          ic: 'shield',
          n: 3,
          desc: 'The conversation looks calm so far. Keep reading carefully.',
        },
        medium: {
          fg: '#F5C84A',
          bd: '#6B5216',
          label: 'MEDIUM RISK',
          ic: 'warn',
          n: 6,
          desc: 'The other side is building trust and creating urgency.',
        },
        high: {
          fg: '#F0853A',
          bd: '#6B3A16',
          label: 'HIGH RISK',
          ic: 'skull',
          n: 9,
          desc: 'Clear signs of fraud have appeared in this conversation.',
        },
        critical: {
          fg: '#F0524B',
          bd: '#6B2029',
          label: 'CRITICAL RISK',
          ic: 'skull',
          n: 12,
          desc: 'Acting now could cost you money or access to your account.',
        },
      }[s.risk] || {}
    const segCol = (i) => (i < 3 ? '#2FE0A0' : i < 6 ? '#F5C84A' : i < 9 ? '#F0853A' : '#F0524B')
    const otherIsSeller = sc.with === 'SELLER'
    const avOther = otherIsSeller ? 'seller' : 'buyer'
    const avYou = otherIsSeller ? 'buyer' : 'seller'
    const CH = [
      { bg: '#07231A', bd: '#1E6B58', fg: '#2FE0A0' },
      { bg: '#0A1C33', bd: '#1E4670', fg: '#8FC9FF' },
      { bg: '#230C0C', bd: '#6B2029', fg: '#F0857B' },
    ]
    const fbTone = s.fb
      ? s.fb.k === 'safe'
        ? {
            bg: '#07231A',
            bd: '#1E6B58',
            bd2: '#12483A',
            fg: '#2FE0A0',
            ic: 'shield',
            title: 'SAFE CHOICE',
            why: 'WHY THIS IS SAFE',
            anim: 'fadeUp 240ms ease-out both',
            xp: 180,
            dlt: '+3',
          }
        : s.fb.k === 'warn'
          ? {
              bg: '#1A1406',
              bd: '#6B5216',
              bd2: '#4A3910',
              fg: '#F5C84A',
              ic: 'warn',
              title: 'INCOMPLETE CHOICE',
              why: 'WHY THIS IS NOT ENOUGH',
              anim: 'fadeUp 240ms ease-out both',
              xp: 120,
              dlt: '+1',
            }
          : {
              bg: '#1A0908',
              bd: '#6B2029',
              bd2: '#4A1614',
              fg: '#F0857B',
              ic: 'skull',
              title: 'RISKY CHOICE',
              why: 'WHY THIS IS RISKY',
              anim: 'shake 320ms steps(5,end) both',
              xp: 40,
              dlt: '0',
            }
      : {}
    const TIPS = [
      {
        t: 'Verify the Seller',
        d: 'How to check profile credibility.',
        why: 'Look at registration date, review count and whether reviews mention real transactions. A profile created days ago with generic praise is worth pausing over.',
      },
      {
        t: 'Secure Payments',
        d: 'Use platform payments or escrow.',
        why: 'Marketplace checkout holds the money until delivery is confirmed. Any method that skips it also skips your protection.',
      },
      {
        t: 'Price Check',
        d: 'How to research fair prices.',
        why: 'Compare three or four active listings for the same item. A price far below all of them is a hook, not a bargain.',
      },
      {
        t: 'Report & Protect',
        d: 'What to do if something feels off.',
        why: 'Use the report button on the listing or chat. Reporting keeps the evidence attached to the account and helps other users.',
      },
    ]
    return {
      flagIcon: px('flag', 16),
      targetIcon2: px('target', 16),
      shieldSm5: px('shield', 16),
      bulbIcon: px('bulb', 16),
      flagRedIcon: px('flagr', 16),
      bookIcon: px('book', 16),
      barsIcon: px('bars', 16),
      clockIcon: px('clock', 18),
      starIcon3: px('star', 16),
      starIcon4: px('star', 18),
      otherAvatar: px(avOther, 30),
      playIcon: px(m.ic, 28),
      playTitle: m.t,
      playDesc: m.d,
      playDif: m.dif,
      playDifBg: d.bg,
      playDifBd: d.bd,
      playDifFg: d.fg,
      stepLabel: 'STEP ' + Math.min(total, s.ci + 1) + ' OF ' + total,
      stepPct: Math.round(Math.min(1, s.ci / total) * 100) + '%',
      objective: sc.obj[Math.min(sc.obj.length - 1, doneDec)],
      riskFg: RISK.fg,
      riskBd: RISK.bd,
      riskLabel: RISK.label,
      riskDesc: RISK.desc,
      riskIcon: px(RISK.ic, 30),
      riskAnim: s.risk === 'critical' ? 'glowPulse 1.6s ease-out 2' : 'none',
      riskSegs: Array.from({ length: 12 }, (_, i) => ({ c: i < RISK.n ? segCol(i) : '#16324C' })),
      riskMark: Math.max(3, (RISK.n / 12) * 100 - 4) + '%',
      playBadges: [
        {
          icon: px('shield', 26),
          label: '18/45',
          fg: '#2FE0A0',
          bd: '#1E6B58',
          open: () => this.setState({ modal: { kind: 'badge', b: this.BADGES[0] } }),
        },
        {
          icon: px('trophy', 26),
          label: '6/24',
          fg: '#F5C84A',
          bd: '#6B5216',
          open: () => this.setState({ modal: { kind: 'badge', b: this.BADGES[2] } }),
        },
        {
          icon: px('fire', 26),
          label: s.streak + ' DAYS',
          fg: '#F0853A',
          bd: '#6B3A16',
          open: () => this.setState({ modal: { kind: 'badge', b: this.BADGES[4] } }),
        },
      ],
      openTips: () => {
        this.beep('click')
        this.setState({
          modal: {
            kind: 'help',
            title: 'Practical tips',
            text: TIPS.map((t) => t.t + ' — ' + t.why).join('\n\n'),
          },
        })
      },
      exitPlay: () => {
        this.beep('click')
        this.exitMission()
      },
      roleCrumb: 'YOU ARE THE ' + sc.role,
      withCrumb: 'CHAT WITH ' + sc.with,
      convId: '#7F3B2A',
      chatRef: this.chatRef,
      msgs: s.msgs.map((msg) => {
        const sys = msg.who === 'sys'
        const mine = msg.who === 'you'
        const sup = msg.who === 'support'
        return {
          isSys: sys,
          isMsg: !sys,
          t: msg.t,
          time: msg.time,
          icon: sys ? px('warn', 16) : null,
          dir: mine ? 'row-reverse' : 'row',
          anim: mine ? 'msgR' : 'msgL',
          who: mine
            ? 'You'
            : sup
              ? 'Support impersonator'
              : sc.with.charAt(0) + sc.with.slice(1).toLowerCase(),
          avatar: px(mine ? avYou : sup ? 'wizard' : avOther, 30),
          avBg: mine ? '#07231A' : sup ? '#170F26' : '#230F0C',
          avBd: mine ? '#1E6B58' : sup ? '#3C2A63' : '#6B2A20',
          bg: mine ? '#0B2B22' : sup ? '#160F26' : '#101E2E',
          bd: mine ? '#1E6B58' : sup ? '#3C2A63' : '#1B3550',
          fg: '#DCE9F7',
          timeFg: mine ? '#5F9E86' : '#5E7794',
          read: mine ? '✓✓' : '',
        }
      }),
      typing: s.typing,
      hasChoices: !!s.dec && !s.fb,
      choices: (s.dec ? s.dec.opts : []).map((o, i) => {
        const c = CH[i],
          selected = s.sel === i
        return {
          t: o.t,
          icon: px('chat', 18),
          bg: selected ? c.bg : c.bg,
          bd: selected ? c.fg : c.bd,
          arrow: c.fg,
          op: s.sel === null || selected ? '1' : '.4',
          cur: s.sel === null ? 'pointer' : 'not-allowed',
          sh: selected ? '0 0 0 2px ' + c.fg : 'none',
          dis: s.sel !== null,
          on: () => this.choose(i),
        }
      }),
      hasFb: !!s.fb,
      fbBg: fbTone.bg,
      fbBd: fbTone.bd,
      fbBd2: fbTone.bd2,
      fbFg: fbTone.fg,
      fbAnim: fbTone.anim,
      fbIcon: s.fb ? px(fbTone.ic, 26) : null,
      fbTitle: fbTone.title,
      fbWhyLabel: fbTone.why,
      fbXp: fbTone.xp,
      fbSkill: s.fb ? s.fb.skill : '',
      fbDelta: fbTone.dlt,
      fbWhat: s.fb ? s.fb.fb.what : '',
      fbWhy: s.fb ? s.fb.fb.why : '',
      fbSign: s.fb ? s.fb.fb.sign : '',
      fbReal: s.fb ? s.fb.fb.real : '',
      onNext: () => this.next(),
      onMore: () => {
        this.beep('click')
        this.setState({
          modal: {
            kind: 'signal',
            sig: { t: s.fb ? s.fb.fb.sign : '', why: s.fb ? s.fb.fb.real : '' },
          },
        })
      },
      sigCount: s.sigs.length,
      hasSigs: s.sigs.length > 0,
      noSigs: s.sigs.length === 0,
      signals: s.sigs.map((g) => ({
        t: g.t,
        s: g.s || '',
        icon: px(g.ic, 20),
        bd: g.tone === 'red' ? '#6B2029' : '#6B5216',
        on: () => {
          this.beep('click')
          this.setState({ modal: { kind: 'signal', sig: g } })
        },
      })),
      tipsCount: TIPS.length,
      playTips: TIPS.map((t) => ({
        t: t.t,
        d: t.d,
        icon: px('bulb', 16),
        on: () => {
          this.beep('click')
          this.setState({ modal: { kind: 'help', title: t.t, text: t.why } })
        },
      })),
      elapsed: this.mmss(s.elapsed),
      gained: s.gained,
    }
  }

  renderVals() {
    const s = this.state,
      px = (n, sz, st) => this.px(n, sz, st)
    const lvl = this.lvl(),
      xpInLvl = s.xp % 1000
    const navDef = [
      ['missions', 'MISSIONS', 'flag'],
      ['scenarios', 'SCENARIOS', 'target'],
      ['progress', 'PROGRESS', 'bars'],
      ['rules', 'RULES', 'book'],
      ['settings', 'SETTINGS', 'gear'],
    ]
    const activeRoute = s.route === 'play' || s.route === 'result' ? 'missions' : s.route
    const list = this.filtered(),
      filtering = this.isFiltering()
    const counts = {}
    Object.keys(this.CAT).forEach(
      (k) => (counts[k] = this.MISSIONS.filter((m) => m.cat === k).length)
    )
    const progCounts = { all: this.MISSIONS.length }
    ;['new', 'progress', 'completed', 'locked'].forEach(
      (k) => (progCounts[k] = this.MISSIONS.filter((m) => this.statusOf(m) === k).length)
    )
    const chipBase = (a) =>
      a
        ? { bg: '#07231A', bd: '#2FE0A0', fg: '#2FE0A0' }
        : { bg: '#0C1826', bd: '#1B3550', fg: '#9FB6D0' }
    const md = this.modalVM()
    const instinctFilled = Math.min(10, Math.max(1, Math.round(lvl)))

    const vw = s.vw || 1600,
      mob = vw < 860
    return {
      motionAttr: s.motion,
      isMobile: mob,
      navDisplay: mob ? 'none' : 'flex',
      stick: mob ? 'static' : 'sticky',
      homeGrid: vw >= 1240 ? 'minmax(0,1fr) 470px' : '1fr',
      hubGrid: vw >= 1080 ? '300px minmax(0,1fr)' : '1fr',
      playGrid:
        vw >= 1400 ? '300px minmax(0,1fr) 300px' : vw >= 1000 ? '300px minmax(0,1fr)' : '1fr',
      duoGrid: vw >= 760 ? '1fr 1fr' : '1fr',
      tipGrid:
        vw >= 1100 ? 'minmax(0,1.05fr) repeat(3,minmax(0,1fr))' : vw >= 700 ? '1fr 1fr' : '1fr',
      heroFs: vw >= 1100 ? '64px' : vw >= 700 ? '48px' : '34px',
      isHome: s.route === 'home',
      isMissions: s.route === 'missions',
      isPlay: s.route === 'play',
      isResult: s.route === 'result',
      isProgress: s.route === 'progress',
      isRules: s.route === 'rules',
      isSettings: s.route === 'settings',
      isScenarios: s.route === 'scenarios',
      logoIcon: px('shield', 30),
      avatarIcon: px('buyer', 32),
      fireIcon: px('fire', 20),
      fireIcon2: px('fire', 18),
      shieldSm: px('shield', 20),
      shieldSm2: px('shield', 18),
      shieldSm3: px('shield', 18),
      shieldSm4: px('shield', 20),
      skullSm: px('skull', 18),
      starIcon: px('star', 18),
      starIcon2: px('star', 20),
      sparkIcon: px('star', 18),
      trophySm: px('trophy', 18),
      trophySm2: px('trophy', 18),
      heartIcon: px('heart', 22),
      medalIcon: px('medal', 36),
      chestIcon: px('chest', 34),
      wizardIcon: px('wizard', 26),
      lockIcon: px('lock', 18),
      targetIcon: px('target', 32),
      catIcon: px('warn', 15),
      progIcon: px('bars', 15),
      sortIcon: px('star', 14),
      emptyIcon: px('glass', 52),
      cityArt: this.city(),
      heroShield: px('shield', 62),
      heroSkull: px('skull', 44),
      xp: s.xp,
      xpInLvl,
      xpPct: Math.round(xpInLvl / 10) + '%',
      lvlPad: (lvl < 10 ? '0' : '') + lvl,
      streak: s.streak,
      badges: s.badgeCount,
      doneCount: s.completed.length,
      inProgress: s.mid && !s.result ? 1 : 0,
      total: this.MISSIONS.length,
      soundOn: s.sound,
      soundLabel: s.sound ? 'ON' : 'OFF',
      soundIcon: px(s.sound ? 'sound' : 'mute', 20),
      soundBd: s.sound ? '#1E6B58' : '#2A5175',
      soundFg: s.sound ? '#2FE0A0' : '#7A93AE',
      toggleSound: () => {
        this.beep('toggle')
        this.set({ sound: !s.sound })
        this.toast(s.sound ? 'Sound muted.' : 'Sound enabled.', 'ok')
      },
      profileOpen: s.profile,
      toggleProfile: () => {
        this.beep('click')
        this.setState({ profile: !s.profile })
      },
      profileItems: [
        {
          label: 'Profile',
          icon: px('wizard', 18),
          color: '#BBD2EA',
          go: () => this.go('progress'),
        },
        {
          label: 'Settings',
          icon: px('gear', 18),
          color: '#BBD2EA',
          go: () => this.go('settings'),
        },
        {
          label: 'Reset demo',
          icon: px('warn', 18),
          color: '#F0857B',
          go: () => {
            this.setState({ profile: false, modal: { kind: 'reset' } })
          },
        },
      ],
      nav: navDef.map((n) => ({
        label: n[1],
        icon: px(n[2], 18, null, activeRoute === n[0] ? '#2FE0A0' : '#7A93AE'),
        go: () => this.go(n[0]),
        cur: activeRoute === n[0] ? 'page' : 'false',
        color: activeRoute === n[0] ? '#2FE0A0' : '#8FA3BD',
        bar: activeRoute === n[0] ? '#2FE0A0' : 'transparent',
      })),
      goHome: () => this.go('home'),
      goMissions: () => this.go('missions'),
      goProgress: () => this.go('progress'),
      goRules: () => this.go('rules'),
      goSettings: () => this.go('settings'),
      goScenarios: () => this.go('scenarios'),
      startTopMission: () => this.startMission('too-good'),

      roleCards: [
        {
          title: 'BUYER',
          aria: 'Train as a buyer',
          desc: 'Learn to spot scams while buying items from others.',
          bg: '#07231A',
          bd: '#1E6B58',
          bd2: '#12483A',
          fg: '#2FE0A0',
          btnBd: '#7CFFD0',
          btnSh: '#0B5B42',
          shadow: '0 0 0 1px rgba(47,224,160,.12)',
          cta: 'START BUYER MISSIONS',
          char: px('buyer', 118),
          icon: px('cart', 24),
          hdrDir: 'row',
          bobDelay: '0s',
          i1: px('shield', 16),
          i2: px('trophy', 16),
          n1: 9,
          n2: 3,
          n3: 240,
          pick: () => {
            this.setState({ fRole: 'BUYER' })
            this.go('missions')
          },
          key: this.keyAct(() => {
            this.setState({ fRole: 'BUYER' })
            this.go('missions')
          }),
        },
        {
          title: 'SELLER',
          aria: 'Train as a seller',
          desc: 'Learn to avoid scams while selling items to others.',
          bg: '#230F0C',
          bd: '#6B2A20',
          bd2: '#4A1C16',
          fg: '#F0524B',
          btnBd: '#FF9088',
          btnSh: '#8C1A18',
          shadow: '0 0 0 1px rgba(240,82,75,.12)',
          cta: 'START SELLER MISSIONS',
          char: px('seller', 118),
          icon: px('store', 24),
          hdrDir: 'row-reverse',
          bobDelay: '.5s',
          i1: px('shield', 16),
          i2: px('trophy', 16),
          n1: 9,
          n2: 3,
          n3: 300,
          pick: () => {
            this.setState({ fRole: 'SELLER' })
            this.go('missions')
          },
          key: this.keyAct(() => {
            this.setState({ fRole: 'SELLER' })
            this.go('missions')
          }),
        },
      ],
      sparks: [
        { l: '11%', t: '16%', c: '#4FD8E8', fs: '20px', dur: '2.2s', delay: '0s' },
        { l: '27%', t: '62%', c: '#2FE0A0', fs: '15px', dur: '1.8s', delay: '.4s' },
        { l: '36%', t: '9%', c: '#3D8BFD', fs: '17px', dur: '2.6s', delay: '.9s' },
        { l: '60%', t: '12%', c: '#4FD8E8', fs: '15px', dur: '2s', delay: '.2s' },
        { l: '73%', t: '44%', c: '#2FE0A0', fs: '21px', dur: '2.4s', delay: '.7s' },
        { l: '89%', t: '18%', c: '#3D8BFD', fs: '16px', dur: '1.9s', delay: '1.1s' },
        { l: '6%', t: '52%', c: '#2FE0A0', fs: '15px', dur: '2.3s', delay: '.5s' },
        { l: '93%', t: '58%', c: '#4FD8E8', fs: '18px', dur: '2.1s', delay: '.3s' },
      ],
      tipText: 'Never share SMS codes or passwords. Real platforms will never ask for them.',
      homeLinks: [
        {
          title: 'Learn the Red Flags',
          sub: 'Know what to watch out for',
          bg: '#230C0C',
          bd: '#6B2029',
          fg: '#F0857B',
          icon: px('flagr', 26),
          go: () => this.go('rules'),
        },
        {
          title: 'Practice Scenarios',
          sub: 'Real conversations, real decisions',
          bg: '#0A1728',
          bd: '#1E4670',
          fg: '#8FC9FF',
          icon: px('chat', 26),
          go: () => this.go('scenarios'),
        },
        {
          title: 'Stay Protected',
          sub: 'Safe shopping starts with smart choices',
          bg: '#07231A',
          bd: '#1E6B58',
          fg: '#2FE0A0',
          icon: px('shield', 26),
          go: () => this.go('progress'),
        },
      ],
      featured: this.MISSIONS.filter((m) => m.feat).map((m) => {
        const d = this.difStyle(m.dif)
        const tile = {
          console: { bg: '#1A1030', bd: '#4A2E7A' },
          truck: { bg: '#231607', bd: '#7A4F22' },
          card: { bg: '#0A1C33', bd: '#1E5FB8' },
          chat: { bg: '#0A1C33', bd: '#1E5FB8' },
        }[m.ic] || { bg: '#08151F', bd: '#1B3550' }
        return {
          t: m.t,
          d: m.d,
          dif: m.dif,
          xp: m.xp,
          difFg: d.fg,
          icon: px(m.ic, 30),
          tileBg: tile.bg,
          tileBd: tile.bd,
          open: () => this.openMission(m.id),
        }
      }),
      weeklyPct: '60%',
      weeklyLabel: '3 / 5',

      q: s.q,
      hasQ: !!s.q,
      sort: s.sort,
      onSearch: (e) => this.setState({ q: e.target.value }),
      clearQ: () => this.setState({ q: '' }),
      onSort: (e) => {
        this.beep('click')
        this.setState({ sort: e.target.value })
      },
      clearFilters: () => {
        this.beep('click')
        this.setState({ q: '', fRole: 'ALL', fDif: [], fCat: null, fProg: 'all', sort: 'dif' })
        this.toast('Filters cleared.', 'ok')
      },
      chipGroups: [
        {
          label: 'ROLE',
          icon: px('cart', 15),
          help: () =>
            this.setState({
              modal: {
                kind: 'help',
                title: 'Role filter',
                text: 'Buyer missions train you while purchasing; seller missions train you while selling. “Both” missions apply to either side of a deal.',
              },
            }),
          items: ['ALL', 'BUYER', 'SELLER'].map((r) => {
            const a = s.fRole === r,
              c = chipBase(a)
            return {
              label: r === 'ALL' ? 'All' : r === 'BUYER' ? 'Buyer' : 'Seller',
              active: a,
              bg: c.bg,
              bd: c.bd,
              fg: c.fg,
              icon: r === 'ALL' ? px('chat', 14) : px(r === 'BUYER' ? 'cart' : 'store', 14),
              on: () => {
                this.beep('click')
                this.setState({ fRole: r })
              },
            }
          }),
        },
        {
          label: 'DIFFICULTY',
          icon: px('star', 14),
          help: () =>
            this.setState({
              modal: {
                kind: 'help',
                title: 'Difficulty',
                text: 'Easy missions show one clear scam signal. Medium missions hide the signal behind a plausible story. Hard missions mix several tactics at once.',
              },
            }),
          items: ['EASY', 'MEDIUM', 'HARD'].map((d) => {
            const a = s.fDif.indexOf(d) >= 0,
              st = this.difStyle(d)
            return {
              label: d.charAt(0) + d.slice(1).toLowerCase(),
              active: a,
              bg: a ? st.bg : '#0C1826',
              bd: a ? st.bd : '#1B3550',
              fg: a ? st.fg : '#9FB6D0',
              icon: null,
              on: () => {
                this.beep('click')
                this.setState({ fDif: a ? s.fDif.filter((x) => x !== d) : s.fDif.concat([d]) })
              },
            }
          }),
        },
      ],
      helpCat: () =>
        this.setState({
          modal: {
            kind: 'help',
            title: 'Scam type',
            text: 'Each mission is built around one family of marketplace fraud. Filter by type to drill the area where you feel least confident.',
          },
        }),
      helpProg: () =>
        this.setState({
          modal: {
            kind: 'help',
            title: 'Progress filter',
            text: 'Track what you have already trained. Locked missions open automatically as your level and completed count grow.',
          },
        }),
      cats: Object.keys(this.CAT).map((k) => {
        const a = s.fCat === k
        return {
          label: this.CAT[k],
          n: counts[k],
          active: a,
          icon: px(
            { pay: 'card', ship: 'truck', fake: 'gift', acc: 'lock', off: 'chat', ret: 'box' }[k],
            15
          ),
          bg: a ? '#07231A' : 'transparent',
          bd: a ? '#2FE0A0' : 'transparent',
          fg: a ? '#2FE0A0' : '#9FB6D0',
          on: () => {
            this.beep('click')
            this.setState({ fCat: a ? null : k })
          },
        }
      }),
      progs: [
        ['all', 'All'],
        ['new', 'Not Started'],
        ['progress', 'In Progress'],
        ['completed', 'Completed'],
        ['locked', 'Locked'],
      ].map((p) => {
        const a = s.fProg === p[0]
        return {
          label: p[1],
          n: progCounts[p[0]],
          active: a,
          bg: a ? '#07231A' : 'transparent',
          bd: a ? '#2FE0A0' : 'transparent',
          fg: a ? '#2FE0A0' : '#9FB6D0',
          on: () => {
            this.beep('click')
            this.setState({ fProg: p[0] })
          },
        }
      }),
      curated: !filtering,
      showResults: filtering,
      hasResults: list.length > 0,
      noResults: list.length === 0,
      resultCount: list.length + (list.length === 1 ? ' mission' : ' missions'),
      results: list.map((m) => this.cardVM(m, 'md')),
      featuredBig: this.MISSIONS.filter((m) => m.feat).map((m) => this.cardVM(m, 'lg')),
      newest: this.MISSIONS.filter((m) => m.new).map((m) => this.cardVM(m, 'md')),
      archive: this.MISSIONS.filter((m) => m.arch).map((m) => this.cardVM(m, 'sm')),
      featuredCount: '3 missions',
      newCount: '4 missions',
      lockedCount: '4 missions',

      instinct: Array.from({ length: 10 }, (_, i) => ({
        c: i < instinctFilled ? '#2FE0A0' : '#16324C',
      })),
      badgeStrip: this.BADGES.slice(0, 4).map((b) => ({
        icon: px(b.ic, 18),
        name: b.name,
        bd: b.earned ? '#6B5216' : '#1B3550',
        open: () => {
          this.beep('click')
          this.setState({ modal: { kind: 'badge', b } })
        },
      })),
      moreBadges: this.BADGES.length - 4,

      ...this.playVals(),
      ...this.resultVals(),
      ...this.pageVals(),

      toast: !!s.toast,
      toastText: s.toast ? s.toast.text : '',
      toastBd: s.toast && s.toast.tone === 'warn' ? '#6B5216' : '#1E6B58',
      toastIcon: px(s.toast && s.toast.tone === 'warn' ? 'warn' : 'check', 20),
      closeToast: () => this.setState({ toast: null }),
      profRef: this.profRef,
      backdropClick: (e) => {
        if (
          e.target === e.currentTarget &&
          (!this.state.modal || this.state.modal.kind !== 'reset')
        )
          this.setState({ modal: null })
      },
      modal: !!md.open,
      modalW: md.w,
      modalBd: md.bd,
      modalTitle: md.title,
      modalIcon: md.icon,
      modalBody: md.body,
      modalActions: md.actions || [],
      closeModal: () => {
        this.beep('click')
        this.setState({ modal: null })
      },
    }
  }
}
